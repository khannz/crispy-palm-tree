package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

const addServiceRequestName = "add service"

// NewServiceInfo ...
type NewServiceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	Healtcheck         ServiceHealthcheck  `json:"Healtcheck" validate:"required"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
	BalanceType        string              `json:"balanceType" validate:"required" example:"rr"`
}

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
// @Router /create-service [post]
func (restAPI *RestAPIstruct) createService(w http.ResponseWriter, r *http.Request) {
	createServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(addServiceRequestName, createServiceUUID, restAPI.balancerFacade.Logging)

	var err error
	bytesFromBuf := readIncomeBytes(r)
	createService := &NewServiceInfo{}

	err = json.Unmarshal(bytesFromBuf, createService)
	if err != nil {
		unmarshallIncomeError(err.Error(),
			createServiceUUID,
			w,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := createService.validateCreateService(); validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, createServiceUUID, w, restAPI.balancerFacade.Logging)
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
			w,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(addServiceRequestName, createServiceUUID, restAPI.balancerFacade.Logging)

	serviceInfo := convertDomainServiceInfoToRestUniversalResponse(nwbServiceInfo, true)

	writeUniversalResponse(serviceInfo,
		addServiceRequestName,
		createServiceUUID,
		w,
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
	return nil
}

func validateServiceBalanceType(balanceType string) error {
	switch balanceType { // maybe range by array is better?
	case "rr":
	case "wrr":
	case "lc":
	case "wlc":
	case "lblc":
	case "sh":
	case "mh":
	case "dh":
	case "fo":
	case "ovf":
	case "lblcr":
	case "sed":
	case "nq":
	default:
		return fmt.Errorf("unknown balance type for service: %v; supported types: rr|wrr|lc|wlc|lblc|sh|mh|dh|fo|ovf|lblcr|sed|nq0", balanceType)
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
