package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const getServicesName = "get services"

// getServices godoc
// @tags load balancer
// @Summary Get services
// @Description Больше, чем балансировщик
// @Produce json
// @Success 200 {object} application.GetAllServicesResponse "If all okay"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /services [get]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) getServices(ginContext *gin.Context) {
	getServicesRequestID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": getServicesRequestID}).Infof("got new %v request", getServicesName)
	nwbServices, err := restAPI.balancerFacade.GetServices(getServicesRequestID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": getServicesRequestID,
		}).Errorf("can't %v, got error: %v", getServicesName, err)
		rError := &UniversalResponse{
			ID:                       getServicesRequestID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "got internal error: " + err.Error(),
		}
		ginContext.JSON(http.StatusInternalServerError, rError)
		return
	}

	allServices := convertDomainServicesInfoToRestUniversalResponseWithState(nwbServices, true)

	var extraInfo string
	if len(allServices) == 0 {
		extraInfo = "No services here"
	}
	getServicesResponse := GetAllServicesResponse{
		ID:                       getServicesRequestID,
		JobCompletedSuccessfully: true,
		AllServices:              allServices,
		ExtraInfo:                extraInfo,
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": getServicesRequestID,
	}).Infof("request %v done", getServicesName)

	ginContext.JSON(http.StatusOK, getServicesResponse)
}
