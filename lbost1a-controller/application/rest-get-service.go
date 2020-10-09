package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const getServiceRequestName = "get service state"

// getService godoc
// @tags load balancer
// @Summary Get service
// @Description Beyond the network balance
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Produce json
// @Success 200 {object} application.Service "If all okay"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal error"
// @Router /service/{addr}/{port} [get]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) getService(ginContext *gin.Context) {
	getServiceRequestID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": getServiceRequestID}).Infof("got new %v request", getServiceRequestName)
	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	// FIXME: validate ip and port
	serviceInfo, err := restAPI.balancerFacade.GetServiceState(ip, port, getServiceRequestID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": getServiceRequestID,
		}).Errorf("can't %v, got error: %v", getServiceRequestName, err)
		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": getServiceRequestID,
	}).Infof("request %v done", getServiceRequestName)

	ginContext.JSON(http.StatusOK, serviceInfo)
}

// func customPortValidationForgetAllServiceStateRequest(sl validator.StructLevel) {
// 	nbi := sl.Current().Interface().(GetServiceStateRequest)
// 	port, err := strconv.Atoi(nbi.ServicePort)
// 	if err != nil {
// 		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
// 	}
// 	if !(port > 0) || !(port < 20000) {
// 		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
// 	}
// }
