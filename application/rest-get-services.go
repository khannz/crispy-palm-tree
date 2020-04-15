package application

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

// GetAllServicesRequest ...
type GetAllServicesRequest struct {
	ID string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
}

// GetAllServicesResponse ...
type GetAllServicesResponse struct {
	ID                       string              `json:"id"`
	JobCompletedSuccessfully bool                `json:"jobCompletedSuccessfully"`
	AllServices              []UniversalResponse `json:"allServices,omitempty"`
	ExtraInfo                string              `json:"extraInfo,omitempty"`
}

// getServices godoc
// @tags Network balance services
// @Summary Get all services
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.GetAllServicesRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.GetAllServicesResponse "If all okay"
// @Failure 400 {object} application.GetAllServicesResponse "Bad request"
// @Failure 500 {object} application.GetAllServicesResponse "Internal error"
// @Router /get-services [post]
func (restAPI *RestAPIstruct) getServices(w http.ResponseWriter, r *http.Request) {
	getRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": getRequestUUID,
	}).Info("got get services request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	newGetServicesRequest := &GetAllServicesRequest{}

	err = json.Unmarshal(bytesFromBuf, newGetServicesRequest)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": getRequestUUID,
		}).Errorf("can't unmarshal income get services request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &GetAllServicesResponse{
			ID:                       getRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income get services request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": getRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	validateError := newGetServicesRequest.validateGetServicesRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": getRequestUUID,
		}).Errorf("validate fail for income get services request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       getRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income get services request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": getRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": getRequestUUID,
	}).Infof("change job uuid from %v to %v", getRequestUUID, newGetServicesRequest.ID)
	getRequestUUID = newGetServicesRequest.ID

	nwbServices, err := restAPI.balancerFacade.GetServices(getRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": getRequestUUID,
		}).Errorf("can't get all nwb services, got error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &GetAllServicesResponse{
			ID:                       getRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't get all nwb services, got internal error: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": getRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}
	allNWBServicesResponse := transformDomainServicesInfoToResponseData(nwbServices)
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": getRequestUUID,
	}).Info("job get services is done")

	var extraInfo string
	if len(allNWBServicesResponse) == 0 {
		extraInfo = "No services here"
	}
	getNwbServicesResponse := GetAllServicesResponse{
		ID:                       getRequestUUID,
		JobCompletedSuccessfully: true,
		AllServices:              allNWBServicesResponse,
		ExtraInfo:                extraInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(getNwbServicesResponse)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": getRequestUUID,
		}).Errorf("can't response by request: %v", err)
	}
}

func transformDomainServicesInfoToResponseData(nwbServices []domain.ServiceInfo) []UniversalResponse {
	UniversalResponses := []UniversalResponse{}
	for _, nwbService := range nwbServices {
		UniversalResponse := UniversalResponse{
			ApplicationServers:       transformDomainApplicationServersToRestApplicationServers(nwbService.ApplicationServers),
			ServiceIP:                nwbService.ServiceIP,
			ServicePort:              nwbService.ServicePort,
			HealthcheckType:          nwbService.HealthcheckType,
			JobCompletedSuccessfully: true,
			ExtraInfo:                transformSliceToString(nwbService.ExtraInfo),
		}
		UniversalResponses = append(UniversalResponses, UniversalResponse)
	}
	return UniversalResponses
}

func transformDomainApplicationServersToRestApplicationServers(domainApplicationServers []domain.ApplicationServer) []ServerApplication {
	applicationServers := []ServerApplication{}
	for _, domainApplicationServer := range domainApplicationServers {
		applicationServer := ServerApplication{
			ServerIP:           domainApplicationServer.ServerIP,
			ServerPort:         domainApplicationServer.ServerPort,
			ServerBashCommands: domainApplicationServer.ServerBashCommands,
		}
		applicationServers = append(applicationServers, applicationServer)
	}
	return applicationServers
}

func transformSliceToString(slice []string) string {
	var resultString string
	for _, el := range slice {
		resultString += "\n" + el
	}
	return resultString
}

func (getAllServicesRequest *GetAllServicesRequest) validateGetServicesRequest() error {
	validate := validator.New()
	err := validate.Struct(getAllServicesRequest)
	if err != nil {
		return err
	}
	return nil
}
