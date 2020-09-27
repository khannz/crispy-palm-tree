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
// @Description Больше, чем балансировщик
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Produce json
// @Success 200 {object} application.UniversalResponseWithStates "If all okay"
// @Failure 400 {object} application.UniversalResponseWithStates "Bad request"
// @Failure 500 {object} application.UniversalResponseWithStates "Internal error"
// @Router /service/{addr}/{port} [get]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) getService(ginContext *gin.Context) {
	getServiceRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": getServiceRequestUUID}).Infof("got new %v request", getServiceRequestName)
	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	// FIXME: validate ip and port
	serviceInfoWithState, err := restAPI.balancerFacade.GetServiceState(ip, port, getServiceRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event uuid": getServiceRequestUUID,
		}).Errorf("can't %v, got error: %v", getServiceRequestName, err)
		rError := &UniversalResponse{
			ID:                       getServiceRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "got internal error: %b" + err.Error(),
		}
		ginContext.JSON(http.StatusInternalServerError, rError)
		return
	}

	convertedServiceInfoWithState := convertDomainServiceInfoToRestUniversalResponseWithStates(serviceInfoWithState, true)

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event uuid": getServiceRequestUUID,
	}).Infof("request %v done", getServiceRequestName)

	ginContext.JSON(http.StatusOK, convertedServiceInfoWithState)
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
