package application

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const addServiceRequestName = "add service"

// createService godoc
// @tags Network balance services
// @Summary Create service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.NewServiceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/create-service [post]
func (restAPI *RestAPIstruct) createService(ginContext *gin.Context) {
	createServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(addServiceRequestName, createServiceUUID, restAPI.balancerFacade.Logging)

	createService := &NewServiceInfo{}

	if err := ginContext.ShouldBindJSON(createService); err != nil {
		unmarshallIncomeError(err.Error(),
			createServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := createService.validateCreateService(); validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, createServiceUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(createServiceUUID, createService.ID, restAPI.balancerFacade.Logging)
	createServiceUUID = createService.ID

	nwbServiceInfo, err := restAPI.balancerFacade.CreateService(createService,
		createServiceUUID)
	if err != nil {
		uscaseFail(addServiceRequestName,
			err.Error(),
			createServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(addServiceRequestName, createServiceUUID, restAPI.balancerFacade.Logging)

	serviceInfo := convertDomainServiceInfoToRestUniversalResponse(nwbServiceInfo, true)

	writeUniversalResponse(serviceInfo,
		addServiceRequestName,
		createServiceUUID,
		ginContext,
		restAPI.balancerFacade.Logging)
}

func (createService *NewServiceInfo) validateCreateService() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationForCreateService, NewServiceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(createService); err != nil {
		return err
	}
	if err := validateServiceBalanceType(createService.BalanceType); err != nil {
		return err
	}
	if err := validateServiceRoutingType(createService.RoutingType); err != nil {
		return err
	}
	return nil
}

func customPortValidationForCreateService(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(NewServiceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
