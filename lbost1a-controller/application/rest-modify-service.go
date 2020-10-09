package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const modifyServiceRequestName = "modify service"

// modifyService godoc
// @tags load balancer
// @Summary Modify service
// @Description Beyond the network balance
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Param incomeJSON body application.Service true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.Service "If all okay"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal error"
// @Router /service/{addr}/{port} [put]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) modifyService(ginContext *gin.Context) {
	modifyServiceID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": modifyServiceID}).Infof("got new %v request", modifyServiceRequestName)
	modifyService := &Service{}

	if err := ginContext.ShouldBindJSON(modifyService); err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": modifyServiceID,
		}).Errorf("can't %v, got error: %v", modifyServiceRequestName, err)
		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}
	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	modifyService.IP = ip
	modifyService.Port = port
	if validateError := modifyService.validateNewService(); validateError != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":   modifyServiceRequestName,
			"event id": modifyServiceID,
		}).Errorf("validate fail for income nwb request: %v", validateError.Error())

		ginContext.String(http.StatusBadRequest, validateError.Error())
		return
	}

	serviceInfo, err := restAPI.balancerFacade.ModifyService(modifyService,
		modifyServiceID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": modifyServiceID,
		}).Errorf("can't %v, got error: %v", modifyServiceRequestName, err.Error())

		ginContext.String(http.StatusInternalServerError, err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":   restAPIlogName,
		"event id": modifyServiceID,
	}).Infof("request %v done", modifyServiceRequestName)
	ginContext.JSON(http.StatusOK, serviceInfo)
}
