package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const removeServiceRequestName = "remove service"

// removeService godoc
// @tags Load balancer
// @Summary Remove nlb service
// @Description Больше, чем балансировщик
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/{addr}/{port} [delete]
// @Security ApiKeyAuth
func (restAPI *RestAPIstruct) removeService(ginContext *gin.Context) {
	removeServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	// TODO: log here. and all code below
	logNewRequest(removeServiceRequestName, removeServiceUUID, restAPI.balancerFacade.Logging)

	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	// FIXME: validate ip and port
	err := restAPI.balancerFacade.RemoveService(ip,
		port,
		removeServiceUUID)
	if err != nil {
		uscaseFail(removeServiceRequestName,
			err.Error(),
			removeServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(removeServiceRequestName, removeServiceUUID, restAPI.balancerFacade.Logging)

	serviceRemoved := UniversalResponse{
		ID:                       removeServiceUUID,
		ServiceIP:                ip,
		ServicePort:              port,
		JobCompletedSuccessfully: true,
		ExtraInfo:                "service removed",
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
