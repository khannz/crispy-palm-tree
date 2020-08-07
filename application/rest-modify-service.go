package application

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const modifyServiceRequestName = "modify service"

// modifyService godoc
// @tags Network balance services
// @Summary Modify service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.ModifyServiceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/create-service [post]
func (restAPI *RestAPIstruct) modifyService(ginContext *gin.Context) {
	modifyServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(modifyServiceRequestName, modifyServiceUUID, restAPI.balancerFacade.Logging)

	modifyService := &ModifyServiceInfo{}

	if err := ginContext.ShouldBindJSON(modifyService); err != nil {
		unmarshallIncomeError(err.Error(),
			modifyServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := modifyService.validatemodifyService(); validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, modifyServiceUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(modifyServiceUUID, modifyService.ID, restAPI.balancerFacade.Logging)
	modifyServiceUUID = modifyService.ID

	nwbServiceInfo, err := restAPI.balancerFacade.ModifyService(modifyService,
		modifyServiceUUID)
	if err != nil {
		uscaseFail(modifyServiceRequestName,
			err.Error(),
			modifyServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(modifyServiceRequestName, modifyServiceUUID, restAPI.balancerFacade.Logging)

	serviceInfo := convertDomainServiceInfoToRestUniversalResponse(nwbServiceInfo, true)

	writeUniversalResponse(serviceInfo,
		modifyServiceRequestName,
		modifyServiceUUID,
		ginContext,
		restAPI.balancerFacade.Logging)
}

func (modifyService *ModifyServiceInfo) validatemodifyService() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationFormodifyService, ModifyServiceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(modifyService); err != nil {
		return err
	}
	if err := validateServiceBalanceType(modifyService.BalanceType); err != nil {
		return err
	}
	if err := validateServiceRoutingType(modifyService.RoutingType); err != nil {
		return err
	}
	return nil
}

func customPortValidationFormodifyService(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(ModifyServiceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
