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
// @Description Beyond the network balance
// @Produce json
// @Success 200 {object} []application.Service "If all okay"
// @Failure 500 {string} error "Internal error"
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

		ginContext.String(http.StatusInternalServerError, "got internal error: "+err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": getServicesRequestID,
	}).Infof("request %v done", getServicesName)

	ginContext.JSON(http.StatusOK, nwbServices)
}
