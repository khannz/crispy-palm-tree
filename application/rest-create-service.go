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
	Healtcheck         ServiceHealtcheck   `json:"Healtcheck" validate:"required"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// createService godoc
// @tags Network balance services
// @Summary Create service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.NewBalanceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /create-service [post]
func (restAPI *RestAPIstruct) createService(w http.ResponseWriter, r *http.Request) {
	createServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": createServiceUUID,
	}).Info("got new add nwb request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	createService := &NewBalanceInfo{}

	err = json.Unmarshal(bytesFromBuf, createService)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": createServiceUUID,
		}).Errorf("can't unmarshal income nwb request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       createServiceUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income nwb request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": createServiceUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": createServiceUUID,
	}).Infof("change job uuid from %v to %v", createServiceUUID, createService.ID)
	createServiceUUID = createService.ID

	_, validateError := createService.validateCreateService()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": createServiceUUID,
		}).Errorf("validate fail for income nwb request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       createServiceUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income nwb request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": createServiceUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	err = restAPI.balancerFacade.CreateService(createService,
		createServiceUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": createServiceUUID,
		}).Errorf("can't create new nwb, got error: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &UniversalResponse{
			ID:                       createServiceUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't create new nwb, got internal error: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": createServiceUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": createServiceUUID,
	}).Info("new nwb created")

	nwbCreated := UniversalResponse{
		ID:                       createServiceUUID,
		ApplicationServers:       createService.ApplicationServers,
		ServiceIP:                createService.ServiceIP,
		ServicePort:              createService.ServicePort,
		HealthcheckType:          "", // FIXME: must be set
		JobCompletedSuccessfully: true,
		ExtraInfo:                "new nwb created",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nwbCreated)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": createServiceUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func (createService *NewBalanceInfo) convertDataForNWBService() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range createService.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (createService *NewBalanceInfo) validateCreateService() (map[string]string, error) {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationForcreateService, NewBalanceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(createService)
	if err != nil {
		return nil, err
	}
	applicationServersMap := createService.convertDataForNWBService()
	err = deepValidateServiceInfo(createService.ServiceIP, createService.ServicePort, applicationServersMap)
	if err != nil {
		return nil, err
	}
	return applicationServersMap, err
}

func customPortValidationForcreateService(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(NewBalanceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
