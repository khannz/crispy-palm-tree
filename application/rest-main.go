package application

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/jwt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	// Need for httpSwagger
	_ "github.com/khannz/crispy-palm-tree/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const restAPIlogName = "restAPI"

// User ...
type User struct {
	login    string
	password string
}

// NewUser ...
func NewUser(login, password string) User {
	return User{
		login:    login,
		password: password,
	}
}

// Authorization ...
type Authorization struct {
	mainSecret        string
	mainRefreshSecret string
	users             []User
}

// NewAuthorization ...
func NewAuthorization(mainSecret, mainRefreshSecret string, users []User) *Authorization {
	return &Authorization{
		mainSecret:        mainSecret,
		mainRefreshSecret: mainRefreshSecret,
		users:             users,
	}
}

// RestAPIstruct restapi entity
type RestAPIstruct struct {
	server         *http.Server
	router         *gin.Engine
	balancerFacade *BalancerFacade
	authorization  *Authorization
}

// NewRestAPIentity ...
func NewRestAPIentity(ip, port string, authorization *Authorization, balancerFacade *BalancerFacade, logger *logrus.Logger) *RestAPIstruct {
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
		authorization:  authorization,
	}

	return restAPI
}

// UpRestAPI ...
func (restAPI *RestAPIstruct) UpRestAPI() {
	service := restAPI.router.Group("/service")
	service.Use(jwt.Auth(restAPI.authorization.mainSecret))
	service.POST("/create-service", restAPI.createService)
	service.POST("/remove-service", restAPI.removeService)
	service.POST("/get-services", restAPI.getServices)
	service.POST("/add-application-servers", restAPI.addApplicationServers)
	service.POST("/remove-application-servers", restAPI.removeApplicationServers)
	service.POST("/get-service", restAPI.getService)
	service.POST("/modify-service", restAPI.modifyService)

	url := ginSwagger.URL("http://" + restAPI.server.Addr + "/swagger/doc.json") // The url pointing to API definition
	restAPI.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	newJWT := restAPI.router.Group("/jwt")
	newJWT.POST("/request-token", restAPI.tokenRequest)

	refreshJWT := restAPI.router.Group("/jwt")
	refreshJWT.POST("/refresh-token", restAPI.tokenRefresh)
	newJWT.Use(jwt.Auth(restAPI.authorization.mainRefreshSecret))

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
