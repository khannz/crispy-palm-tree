package application

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const addApplicationServersRequestName = "add application servers"

// addApplicationServers godoc
// @tags Network balance services
// @Summary Add application servers
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.AddApplicationServersRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/add-application-servers [post]
func (restAPI *RestAPIstruct) addApplicationServers(ginContext *gin.Context) {
	addApplicationServersRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(addApplicationServersRequestName, addApplicationServersRequestUUID, restAPI.balancerFacade.Logging)

	addApplicationServersRequest := &AddApplicationServersRequest{}

	if err := ginContext.ShouldBindJSON(addApplicationServersRequest); err != nil {
		unmarshallIncomeError(err.Error(),
			addApplicationServersRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := addApplicationServersRequest.validateAddApplicationServersRequest(); validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, addApplicationServersRequestUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(addApplicationServersRequestUUID, addApplicationServersRequest.ID, restAPI.balancerFacade.Logging)
	addApplicationServersRequestUUID = addApplicationServersRequest.ID

	updatedServiceInfo, err := restAPI.balancerFacade.AddApplicationServers(addApplicationServersRequest,
		addApplicationServersRequestUUID)
	if err != nil {
		uscaseFail(addApplicationServersRequestName,
			err.Error(),
			addApplicationServersRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(addApplicationServersRequestName, addApplicationServersRequestUUID, restAPI.balancerFacade.Logging)

	convertedServiceInfo := convertDomainServiceInfoToRestUniversalResponse(updatedServiceInfo, true)
	writeUniversalResponse(convertedServiceInfo,
		addApplicationServersRequestName,
		addApplicationServersRequestUUID,
		ginContext,
		restAPI.balancerFacade.Logging)
}

func (addApplicationServersRequest *AddApplicationServersRequest) convertDataAddApplicationServersRequest() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range addApplicationServersRequest.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (addApplicationServersRequest *AddApplicationServersRequest) validateAddApplicationServersRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortAddApplicationServersRequestValidation, AddApplicationServersRequest{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(addApplicationServersRequest); err != nil {
		return err
	}
	return nil
}

func customPortAddApplicationServersRequestValidation(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(AddApplicationServersRequest)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
