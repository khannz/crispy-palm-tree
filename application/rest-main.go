package application

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	// Need for httpSwagger
	_ "github.com/khannz/crispy-palm-tree/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const restAPIlogName = "restAPI"

// @title NLB service swagger
// @version 1.0.1
// @description create/delete nlb service.
// @Tags New nlb service
// @tag.name Link for docs
// @tag.docs.url http://kb.sdn.sbrf.ru/display/SDN/*
// @tag.docs.description Docs at confluence
// @contact.name Ivan Tikhonov
// @contact.email sdn@sberbank.ru

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// RestAPIstruct restapi entity
type RestAPIstruct struct {
	server         *http.Server
	router         *gin.Engine
	balancerFacade *BalancerFacade
}

// NewRestAPIentity ...
func NewRestAPIentity(ip, port string, balancerFacade *BalancerFacade, logger *logrus.Logger) *RestAPIstruct { // TODO: authentication (Oauth2?)
	router := gin.Default()
	router.Use(ginrus.Ginrus(logger, time.RFC3339, false))
	fullAddres := ip + ":" + port
	server := &http.Server{
		Addr: fullAddres, // ip + ":" + port - not working here
		// Good practice to set timeouts to avoid Slowloris attacks.
		// WriteTimeout: time.Second * 15,
		// ReadTimeout:  time.Second * 15,
		// IdleTimeout:  time.Second * 60,
		Handler: router,
	}

	restAPI := &RestAPIstruct{
		server:         server,
		router:         router,
		balancerFacade: balancerFacade,
	}

	return restAPI
}

// UpRestAPI ...
func (restAPI *RestAPIstruct) UpRestAPI() {
	restAPI.router.POST("/create-service", restAPI.createService)
	restAPI.router.POST("/remove-service", restAPI.removeService)
	restAPI.router.POST("/get-services", restAPI.getServices)
	restAPI.router.POST("/add-application-servers", restAPI.addApplicationServers)
	restAPI.router.POST("/remove-application-servers", restAPI.removeApplicationServers)
	restAPI.router.POST("/get-service", restAPI.getService)
	restAPI.router.POST("/modify-service", restAPI.modifyService)

	url := ginSwagger.URL("http://" + restAPI.server.Addr + "/swagger/doc.json") // The url pointing to API definition
	restAPI.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// restAPI.router.PathPrefix("/swagger-ui.html/").Handler(httpSwagger.WrapHandler)

	err := restAPI.server.ListenAndServe()
	if err != nil {
		restAPI.balancerFacade.Logging.Infof("rest api down: %v", err)
	}
}

// GracefulShutdownRestAPI ...
func (restAPI *RestAPIstruct) GracefulShutdownRestAPI(gracefulShutdownCommandForRestAPI, restAPIisDone chan struct{}) {
	<-gracefulShutdownCommandForRestAPI
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity": restAPIlogName,
	}).Info("stoping http server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20*time.Second))
	defer cancel()

	err := restAPI.server.Shutdown(ctx)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity": restAPIlogName,
		}).Errorf("shutdown request error: %v", err)
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity": restAPIlogName,
	}).Info("rest api stoped")

	restAPIisDone <- struct{}{}
}
