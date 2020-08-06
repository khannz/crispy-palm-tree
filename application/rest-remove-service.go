package application

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const removeServiceRequestName = "remove service"

// RemoveServiceInfo ...
type RemoveServiceInfo struct {
	ID          string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP   string `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort string `json:"servicePort" validate:"required" example:"1111"`
}

// removeService godoc
// @tags Network balance services
// @Summary Remove nlb service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.RemoveServiceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /remove-service [post]
func (restAPI *RestAPIstruct) removeService(ginContext *gin.Context) {
	removeServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(removeServiceRequestName, removeServiceUUID, restAPI.balancerFacade.Logging)

	var err error
	bytesFromBuf := readIncomeBytes(ginContext.Request)

	removeService := &RemoveServiceInfo{}

	err = json.Unmarshal(bytesFromBuf, removeService)
	if err != nil {
		unmarshallIncomeError(err.Error(),
			removeServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	validateError := removeService.validateRemoveNWBRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, removeServiceUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(removeServiceUUID, removeService.ID, restAPI.balancerFacade.Logging)
	removeServiceUUID = removeService.ID

	err = restAPI.balancerFacade.RemoveService(removeService,
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
		ServiceIP:                removeService.ServiceIP,
		ServicePort:              removeService.ServicePort,
		JobCompletedSuccessfully: true,
		ExtraInfo:                "service removed",
	}
	writeUniversalResponse(serviceRemoved,
		removeServiceRequestName,
		removeServiceUUID,
		ginContext,
		restAPI.balancerFacade.Logging)
}

func (removeService *RemoveServiceInfo) validateRemoveNWBRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortRemoveServiceInfoValidation, RemoveServiceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(removeService)
	if err != nil {
		return err
	}
	return nil
}

func customPortRemoveServiceInfoValidation(sl validator.StructLevel) {
	nrbi := sl.Current().Interface().(RemoveServiceInfo)
	port, err := strconv.Atoi(nrbi.ServicePort)
	if err != nil {
		sl.ReportError(nrbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nrbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
