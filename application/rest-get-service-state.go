package application

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const getServiceStateRequestName = "get service state"

// GetServiceStateRequest ...
type GetServiceStateRequest struct {
	ID          string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP   string `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort string `json:"servicePort" validate:"required" example:"1111"`
}

// getServiceState godoc
// @tags Network balance services
// @Summary Get service state
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.GetServiceStateRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponseWithStates "If all okay"
// @Failure 400 {object} application.UniversalResponseWithStates "Bad request"
// @Failure 500 {object} application.UniversalResponseWithStates "Internal error"
// @Router /get-service-state [post]
func (restAPI *RestAPIstruct) getServiceState(w http.ResponseWriter, r *http.Request) {
	getServicesRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(getServiceStateRequestName, getServicesRequestUUID, restAPI.balancerFacade.Logging)

	var err error
	bytesFromBuf := readIncomeBytes(r)

	newGetServiceStateRequest := &GetServiceStateRequest{}

	err = json.Unmarshal(bytesFromBuf, newGetServiceStateRequest)
	if err != nil {
		unmarshallIncomeError(err.Error(),
			getServicesRequestUUID,
			w,
			restAPI.balancerFacade.Logging)
		return
	}

	validateError := newGetServiceStateRequest.validateGetServiceStateRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, getServicesRequestUUID, w, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(getServicesRequestUUID, newGetServiceStateRequest.ID, restAPI.balancerFacade.Logging)
	getServicesRequestUUID = newGetServiceStateRequest.ID

	serviceInfoWithState, err := restAPI.balancerFacade.GetServiceState(newGetServiceStateRequest)
	if err != nil {
		uscaseFail(getServiceStateRequestName,
			err.Error(),
			getServicesRequestUUID,
			w,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(getServiceStateRequestName, getServicesRequestUUID, restAPI.balancerFacade.Logging)
	convertedServiceInfoWithState := convertDomainServiceInfoToRestUniversalResponseWithStates(serviceInfoWithState, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(convertedServiceInfoWithState)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": getServicesRequestUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func (getAllServiceStateRequest *GetServiceStateRequest) validateGetServiceStateRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationForgetAllServiceStateRequest, GetServiceStateRequest{})
	if err := validate.Struct(getAllServiceStateRequest); err != nil {
		return err
	}
	return nil
}

func customPortValidationForgetAllServiceStateRequest(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(GetServiceStateRequest)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
