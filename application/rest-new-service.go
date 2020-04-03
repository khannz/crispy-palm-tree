package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// NewBalanceInfo ...
type NewBalanceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	HealthcheckType    string              `json:"healthcheckType,omitempty"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// newNWBRequest godoc
// @tags Network balance services
// @Summary Create nlb service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.NewBalanceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /newnetworkbalance [post]
func (restAPI *RestAPIstruct) newNWBRequest(w http.ResponseWriter, r *http.Request) {
	newNWBRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newNWBRequestUUID,
	}).Info("got new add nwb request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	newNWBRequest := &NewBalanceInfo{}

	err = json.Unmarshal(bytesFromBuf, newNWBRequest)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newNWBRequestUUID,
		}).Errorf("can't unmarshal income nwb request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       newNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income nwb request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": newNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newNWBRequestUUID,
	}).Infof("change job uuid from %v to %v", newNWBRequestUUID, newNWBRequest.ID)
	newNWBRequestUUID = newNWBRequest.ID

	validateError := newNWBRequest.validateNewNWBRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newNWBRequestUUID,
		}).Errorf("validate fail for income nwb request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       newNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income nwb request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": newNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}
	applicationServersMap := newNWBRequest.convertDataForNWBService()
	err = restAPI.balancerFacade.NewNWBService(newNWBRequest.ServiceIP,
		newNWBRequest.ServicePort,
		applicationServersMap,
		newNWBRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newNWBRequestUUID,
		}).Errorf("can't create new nwb, got error: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &UniversalResponse{
			ID:                       newNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't create new nwb, got internal error: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": newNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newNWBRequestUUID,
	}).Info("new nwb created")

	nwbCreated := UniversalResponse{
		ID:                       newNWBRequestUUID,
		ApplicationServers:       newNWBRequest.ApplicationServers,
		ServiceIP:                newNWBRequest.ServiceIP,
		ServicePort:              newNWBRequest.ServicePort,
		HealthcheckType:          newNWBRequest.HealthcheckType,
		JobCompletedSuccessfully: true,
		ExtraInfo:                "new nwb created",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nwbCreated)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newNWBRequestUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func (newNWBRequest *NewBalanceInfo) convertDataForNWBService() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range newNWBRequest.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (newNWBRequest *NewBalanceInfo) validateNewNWBRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationForNewNWBRequest, NewBalanceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(newNWBRequest)
	if err != nil {
		return err
	}
	return nil
}

func customPortValidationForNewNWBRequest(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(NewBalanceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
