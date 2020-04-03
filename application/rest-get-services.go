package application

import (
	"bytes"
	"encoding/json"
	"net/http"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"github.com/go-playground/validator/v10"
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

// getNWBServices godoc
// @tags Network balance services
// @Summary Get nlb services
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.GetAllServicesRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.GetAllServicesResponse "If all okay"
// @Failure 400 {object} application.GetAllServicesResponse "Bad request"
// @Failure 500 {object} application.GetAllServicesResponse "Internal error"
// @Router /networkservicesinfo [post]
func (restAPI *RestAPIstruct) getNWBServices(w http.ResponseWriter, r *http.Request) {
	newGetNWBRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newGetNWBRequestUUID,
	}).Info("got new get nwb request")

	var err error
	buf := new(bytes.Buffer) // read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	newGetNWBRequest := &GetAllServicesRequest{}

	err = json.Unmarshal(bytesFromBuf, newGetNWBRequest)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newGetNWBRequestUUID,
		}).Errorf("can't unmarshal income get nwb request: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &GetAllServicesResponse{
			ID:                       newGetNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't unmarshal income get nwb request: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": newGetNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	validateError := newGetNWBRequest.validateNewGetNWBRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newGetNWBRequestUUID,
		}).Errorf("validate fail for income get request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &UniversalResponse{
			ID:                       newGetNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't validate income get nwb request: " + stringValidateError,
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": newGetNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newGetNWBRequestUUID,
	}).Infof("change job uuid from %v to %v", newGetNWBRequestUUID, newGetNWBRequest.ID)
	newGetNWBRequestUUID = newGetNWBRequest.ID

	nwbServices, err := restAPI.balancerFacade.GetNWBServices(newGetNWBRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newGetNWBRequestUUID,
		}).Errorf("can't get all nwb services, got error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &GetAllServicesResponse{
			ID:                       newGetNWBRequestUUID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "can't get all nwb services, got internal error: " + err.Error(),
		}
		err := json.NewEncoder(w).Encode(rError)
		if err != nil {
			restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
				"entity":     restAPIlogName,
				"event uuid": newGetNWBRequestUUID,
			}).Errorf("can't response by request: %v", err)
		}
		return
	}
	allNWBServicesResponse := transformDomainServicesInfoToResponseData(nwbServices)
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newGetNWBRequestUUID,
	}).Info("job get services is done")

	var extraInfo string
	if len(allNWBServicesResponse) == 0 {
		extraInfo = "No services here"
	}
	getNwbServicesResponse := GetAllServicesResponse{
		ID:                       newGetNWBRequestUUID,
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
			"event uuid": newGetNWBRequestUUID,
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

func (getAllServicesRequest *GetAllServicesRequest) validateNewGetNWBRequest() error {
	validate := validator.New()
	err := validate.Struct(getAllServicesRequest)
	if err != nil {
		return err
	}
	return nil
}
