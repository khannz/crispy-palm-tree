package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const removeServiceRequestName = "remove service"

// removeService godoc
// @tags load balancer
// @Summary Remove service
// @Description Beyond the network balance
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Produce json
// @Success 200 {object} application.Service "If all okay"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal error"
// @Router /service/{addr}/{port} [delete]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) removeService(ginContext *gin.Context) {
	removeServiceID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": removeServiceID}).Infof("got new %v request", removeServiceRequestName)
	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	// FIXME: validate ip and port
	err := restAPI.balancerFacade.RemoveService(ip,
		port,
		removeServiceID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": removeServiceID,
		}).Errorf("can't %v, got error: %v", removeServiceID, err)

		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": removeServiceID,
	}).Infof("request %v done", removeServiceRequestName)

	serviceRemoved := &Service{
		IP:   ip,
		Port: port,
	}
	ginContext.JSON(http.StatusOK, serviceRemoved)
}

// func (removeService *RemoveServiceInfo) validateRemoveNWBRequest() error {
// 	validate := validator.New()
// 	validate.RegisterStructValidation(customPortRemoveServiceInfoValidation, RemoveServiceInfo{})
// 	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
// 	if err := validate.Struct(removeService); err != nil {
// 		return modifyValidateError(err)
// 	}
// 	return nil
// }

// func customPortRemoveServiceInfoValidation(sl validator.StructLevel) {
// 	nrbi := sl.Current().Interface().(RemoveServiceInfo)
// 	port, err := strconv.Atoi(nrbi.ServicePort)
// 	if err != nil {
// 		sl.ReportError(nrbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
// 	}
// 	if !(port > 0) || !(port < 20000) {
// 		sl.ReportError(nrbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
// 	}
// }
