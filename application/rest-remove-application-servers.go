package application

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const removeApplicationServersRequestName = "remove application servers"

// removeApplicationServers godoc
// @tags load balancer
// @Summary Remove application servers
// @Description Больше, чем балансировщик
// @Param incomeJSON body application.RemoveApplicationServersRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/remove-application-servers [post]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) removeApplicationServers(ginContext *gin.Context) {
	removeApplicationServersRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	logNewRequest(removeApplicationServersRequestName, removeApplicationServersRequestUUID, restAPI.balancerFacade.Logging)

	removeApplicationServersRequest := &RemoveApplicationServersRequest{}

	if err := ginContext.ShouldBindJSON(removeApplicationServersRequest); err != nil {
		unmarshallIncomeError(err.Error(),
			removeApplicationServersRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := removeApplicationServersRequest.validateRemoveApplicationServersRequest(); validateError != nil {
		validateIncomeError(validateError.Error(), removeApplicationServersRequestUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(removeApplicationServersRequestUUID, removeApplicationServersRequest.ID, restAPI.balancerFacade.Logging)
	removeApplicationServersRequestUUID = removeApplicationServersRequest.ID

	updatedServiceInfo, err := restAPI.balancerFacade.RemoveApplicationServers(removeApplicationServersRequest,
		removeApplicationServersRequestUUID)
	if err != nil {
		uscaseFail(removeApplicationServersRequestName,
			err.Error(),
			removeApplicationServersRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(removeApplicationServersRequestName, removeApplicationServersRequestUUID, restAPI.balancerFacade.Logging)

	convertedServiceInfo := convertDomainServiceInfoToRestUniversalResponse(updatedServiceInfo, true)
	ginContext.JSON(http.StatusOK, convertedServiceInfo)
}

func (removeApplicationServersRequest *RemoveApplicationServersRequest) validateRemoveApplicationServersRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortRemoveApplicationServersRequestValidation, RemoveApplicationServersRequest{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	if err := validate.Struct(removeApplicationServersRequest); err != nil {
		return modifyValidateError(err)
	}
	return nil
}

func customPortRemoveApplicationServersRequestValidation(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(RemoveApplicationServersRequest)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
