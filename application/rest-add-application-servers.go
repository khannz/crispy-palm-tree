package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// AddApplicationServersRequest ...
type AddApplicationServersRequest struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	Healtcheck         ServiceHealtcheck   `json:"Healtcheck" validate:"required"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

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
// @Router /add-application-servers [post]
func (restAPI *RestAPIstruct) addApplicationServers(w http.ResponseWriter, r *http.Request) {
	addApplicationServersRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": addApplicationServersRequestUUID,
	}).Info("got new add application servers request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	addApplicationServersRequest := &AddApplicationServersRequest{}

	err = json.Unmarshal(bytesFromBuf, addApplicationServersRequest)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": addApplicationServersRequestUUID,
		}).Errorf("can't unmarshal income add application servers request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       addApplicationServersRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income add application servers request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": addApplicationServersRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	_, validateError := addApplicationServersRequest.validateAddApplicationServersRequest() // refactor, map not needed
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": addApplicationServersRequestUUID,
		}).Errorf("validate fail for income add application servers request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       addApplicationServersRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income add application servers nwb request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": addApplicationServersRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": addApplicationServersRequestUUID,
	}).Infof("change job uuid from %v to %v", addApplicationServersRequestUUID, addApplicationServersRequest.ID)
	addApplicationServersRequestUUID = addApplicationServersRequest.ID

	updatedServiceInfo, err := restAPI.balancerFacade.AddApplicationServers(addApplicationServersRequest,
		addApplicationServersRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": addApplicationServersRequestUUID,
		}).Errorf("can't add application servers, got error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		updatedServiceInfo.ExtraInfo = []string{"can't add application servers, got internal error: " + err.Error()}
		rError := transformDomainServiceInfoToResponseData(updatedServiceInfo, false)
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": addApplicationServersRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": addApplicationServersRequestUUID,
	}).Info("job add application servers is done")

	convertedServiceInfo := transformDomainServiceInfoToResponseData(updatedServiceInfo, true)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(convertedServiceInfo)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": addApplicationServersRequestUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func (addApplicationServersRequest *AddApplicationServersRequest) convertDataAddApplicationServersRequest() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range addApplicationServersRequest.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (addApplicationServersRequest *AddApplicationServersRequest) validateAddApplicationServersRequest() (map[string]string, error) {
	validate := validator.New()
	validate.RegisterStructValidation(customPortAddApplicationServersRequestValidation, AddApplicationServersRequest{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(addApplicationServersRequest)
	if err != nil {
		return nil, err
	}
	applicationServersMap := addApplicationServersRequest.convertDataAddApplicationServersRequest()
	err = deepValidateServiceInfo(addApplicationServersRequest.ServiceIP, addApplicationServersRequest.ServicePort, applicationServersMap)
	if err != nil {
		return nil, err
	}
	return applicationServersMap, nil
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
