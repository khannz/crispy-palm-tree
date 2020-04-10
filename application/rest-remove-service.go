package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// RemoveBalanceInfo ...
type RemoveBalanceInfo struct {
	ID          string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP   string `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort string `json:"servicePort" validate:"required" example:"1111"`
}

// removeNWBRequest godoc
// @tags Network balance services
// @Summary Remove nlb service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.RemoveBalanceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /remove-service [post]
func (restAPI *RestAPIstruct) removeNWBRequest(w http.ResponseWriter, r *http.Request) {
	removeNWBRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeNWBRequestUUID,
	}).Info("got new remove nwb request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	removeNWBRequest := &RemoveBalanceInfo{}

	err = json.Unmarshal(bytesFromBuf, removeNWBRequest)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeNWBRequestUUID,
		}).Errorf("can't unmarshal income remove nwb request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       removeNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income nwb request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": removeNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeNWBRequestUUID,
	}).Infof("change job uuid from %v to %v", removeNWBRequestUUID, removeNWBRequest.ID)
	removeNWBRequestUUID = removeNWBRequest.ID

	validateError := removeNWBRequest.validateRemoveNWBRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeNWBRequestUUID,
		}).Errorf("validate fail for income remove nwb request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       removeNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income nwb request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": removeNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	err = restAPI.balancerFacade.RemoveService(removeNWBRequest.ServiceIP,
		removeNWBRequest.ServicePort,
		removeNWBRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeNWBRequestUUID,
		}).Errorf("can't remove old nwb, got error: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &UniversalResponse{
			ID:                       removeNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't create new nwb, got internal error: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": removeNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeNWBRequestUUID,
	}).Info("nwb removed")

	nwbRemoved := UniversalResponse{
		ID:                       removeNWBRequestUUID,
		ServiceIP:                removeNWBRequest.ServiceIP,
		ServicePort:              removeNWBRequest.ServicePort,
		JobCompletedSuccessfully: true,
		ExtraInfo:                "nwb removed",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nwbRemoved)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeNWBRequestUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func (removeNWBRequest *RemoveBalanceInfo) validateRemoveNWBRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortRemoveBalanceInfoValidation, RemoveBalanceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(removeNWBRequest)
	if err != nil {
		return err
	}
	return nil
}

func customPortRemoveBalanceInfoValidation(sl validator.StructLevel) {
	nrbi := sl.Current().Interface().(RemoveBalanceInfo)
	port, err := strconv.Atoi(nrbi.ServicePort)
	if err != nil {
		sl.ReportError(nrbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nrbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
