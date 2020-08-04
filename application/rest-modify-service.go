package application

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

const modifyServiceRequestName = "modify service"

// ModifyServiceInfo ...
type ModifyServiceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	Healtcheck         ServiceHealthcheck  `json:"Healtcheck" validate:"required"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
	BalanceType        string              `json:"balanceType" validate:"required" example:"rr"`
	RoutingType        string              `json:"routingType" validate:"required" example:"masquarading,tunneling"`
}

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
// @Router /create-service [post]
func (restAPI *RestAPIstruct) modifyService(w http.ResponseWriter, r *http.Request) {
	modifyServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(modifyServiceRequestName, modifyServiceUUID, restAPI.balancerFacade.Logging)

	var err error
	bytesFromBuf := readIncomeBytes(r)
	modifyService := &ModifyServiceInfo{}

	err = json.Unmarshal(bytesFromBuf, modifyService)
	if err != nil {
		unmarshallIncomeError(err.Error(),
			modifyServiceUUID,
			w,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := modifyService.validatemodifyService(); validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, modifyServiceUUID, w, restAPI.balancerFacade.Logging)
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
			w,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(modifyServiceRequestName, modifyServiceUUID, restAPI.balancerFacade.Logging)

	serviceInfo := convertDomainServiceInfoToRestUniversalResponse(nwbServiceInfo, true)

	writeUniversalResponse(serviceInfo,
		modifyServiceRequestName,
		modifyServiceUUID,
		w,
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
