package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// RemoveApplicationServersRequest ...
type RemoveApplicationServersRequest struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// removeApplicationServers godoc
// @tags Network balance services
// @Summary Remove application servers
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.RemoveApplicationServersRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /remove-application-servers [post]
func (restAPI *RestAPIstruct) removeApplicationServers(w http.ResponseWriter, r *http.Request) {
	removeApplicationServersRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeApplicationServersRequestUUID,
	}).Info("got new remove application servers request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	removeApplicationServersRequest := &RemoveApplicationServersRequest{}

	err = json.Unmarshal(bytesFromBuf, removeApplicationServersRequest)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeApplicationServersRequestUUID,
		}).Errorf("can't unmarshal income remove application servers request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       removeApplicationServersRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income remove application servers request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": removeApplicationServersRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	applicationServersMap, validateError := removeApplicationServersRequest.validateRemoveApplicationServersRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeApplicationServersRequestUUID,
		}).Errorf("validate fail for income remove application servers request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       removeApplicationServersRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income remove application servers nwb request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": removeApplicationServersRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeApplicationServersRequestUUID,
	}).Infof("change job uuid from %v to %v", removeApplicationServersRequestUUID, removeApplicationServersRequest.ID)
	removeApplicationServersRequestUUID = removeApplicationServersRequest.ID

	updatedServiceInfo, err := restAPI.balancerFacade.RemoveApplicationServers(removeApplicationServersRequest.ServiceIP,
		removeApplicationServersRequest.ServicePort,
		applicationServersMap,
		removeApplicationServersRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeApplicationServersRequestUUID,
		}).Errorf("can't remove application servers, got error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		updatedServiceInfo.ExtraInfo = []string{"can't remove application servers, got internal error: " + err.Error()}
		rError := transformDomainServiceInfoToResponseData(updatedServiceInfo, false)
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": removeApplicationServersRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeApplicationServersRequestUUID,
	}).Info("job remove application servers is done")

	convertedServiceInfo := transformDomainServiceInfoToResponseData(updatedServiceInfo, true)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(convertedServiceInfo)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeApplicationServersRequestUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func (removeApplicationServersRequest *RemoveApplicationServersRequest) convertDataRemoveApplicationServersRequest() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range removeApplicationServersRequest.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (removeApplicationServersRequest *RemoveApplicationServersRequest) validateRemoveApplicationServersRequest() (map[string]string, error) {
	validate := validator.New()
	validate.RegisterStructValidation(customPortRemoveApplicationServersRequestValidation, RemoveApplicationServersRequest{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(removeApplicationServersRequest)
	if err != nil {
		return nil, err
	}
	applicationServersMap := removeApplicationServersRequest.convertDataRemoveApplicationServersRequest()
	err = deepValidateServiceInfo(removeApplicationServersRequest.ServiceIP, removeApplicationServersRequest.ServicePort, applicationServersMap)
	if err != nil {
		return nil, err
	}
	return applicationServersMap, nil
}

func customPortRemoveApplicationServersRequestValidation(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(RemoveApplicationServersRequest)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
