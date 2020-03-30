package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	// Need for httpSwagger
	_ "git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/docs"
	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

const restAPIlogName = "restAPI"

// @title NLB service swagger
// @version 1.0.1
// @description create/delete nlb service.
// @Tags New nlb service
// @tag.name Link for docs
// @tag.docs.url http://kb.sdn.sbrf.ru/display/SDN/*
// @tag.docs.description Docs at confluence
// @contact.name Ivan Tikhonov
// @contact.email sdn@sberbank.ru

// ServerApplication ...
type ServerApplication struct {
	ServerIP           string `json:"ip" validate:"required,ipv4" example:"1.1.1.1"`
	ServerPort         string `json:"port" validate:"required" example:"1111"`
	ServerBashCommands string `json:"bashCommands,omitempty" swaggerignore:"true"`
}

// UniversalResponse ...
type UniversalResponse struct {
	ID                       string              `json:"id,omitempty"`
	ApplicationServers       []ServerApplication `json:"applicationServers,omitempty"`
	ServiceIP                string              `json:"serviceIP,omitempty"`
	ServicePort              string              `json:"servicePort,omitempty"`
	HealthcheckType          string              `json:"healthcheckType,omitempty"`
	JobCompletedSuccessfully bool                `json:"jobCompletedSuccessfully"`
	ExtraInfo                string              `json:"extraInfo,omitempty"`
}

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

// NewBalanceInfo ...
type NewBalanceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	HealthcheckType    string              `json:"healthcheckType,omitempty"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// RemoveBalanceInfo ...
type RemoveBalanceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	HealthcheckType    string              `json:"healthcheckType,omitempty"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// RestAPIstruct restapi entity
type RestAPIstruct struct {
	server         *http.Server
	router         *mux.Router
	balancerFacade *BalancerFacade
}

// NewRestAPIentity ...
func NewRestAPIentity(ip, port string, balancerFacade *BalancerFacade) *RestAPIstruct { // TODO: authentication (Oauth2?)
	router := mux.NewRouter()
	fullAddres := ip + ":" + port
	server := &http.Server{
		Addr: fullAddres, // ip + ":" + port - not working here
		// Good practice to set timeouts to avoid Slowloris attacks.
		// WriteTimeout: time.Second * 15,
		// ReadTimeout:  time.Second * 15,
		// IdleTimeout:  time.Second * 60,
		Handler: router,
	}

	restAPI := &RestAPIstruct{
		server:         server,
		router:         router,
		balancerFacade: balancerFacade,
	}

	return restAPI
}

// UpRestAPI ...
func (restAPI *RestAPIstruct) UpRestAPI() {
	restAPI.router.HandleFunc("/newnetworkbalance", restAPI.newNWBRequest).Methods("POST")
	restAPI.router.HandleFunc("/removenetworkbalance", restAPI.removeNWBRequest).Methods("POST")
	restAPI.router.HandleFunc("/networkservicesinfo", restAPI.getNWBServices).Methods("POST")
	restAPI.router.PathPrefix("/swagger-ui.html/").Handler(httpSwagger.WrapHandler)

	err := restAPI.server.ListenAndServe()
	if err != nil {
		restAPI.balancerFacade.Logging.Infof("rest api down: %v", err)
	}
}

////
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

	newGetNWBRequest := &GetAllServicesRequest{} // maybe use pointer here?

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

////
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

	newNWBRequest := &NewBalanceInfo{} // maybe use pointer here?

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

	validateError := newNWBRequest.validateIncomeRequest()
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
// @Router /removenetworkbalance [post]
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

	removeNWBRequest := &RemoveBalanceInfo{} // maybe use pointer here?

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

	validateError := removeNWBRequest.validateIncomeRequest()
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
	applicationServersMap := removeNWBRequest.convertDataForNWBService()
	err = restAPI.balancerFacade.RemoveNWBService(removeNWBRequest.ServiceIP,
		removeNWBRequest.ServicePort,
		applicationServersMap,
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
		ApplicationServers:       removeNWBRequest.ApplicationServers,
		ServiceIP:                removeNWBRequest.ServiceIP,
		ServicePort:              removeNWBRequest.ServicePort,
		HealthcheckType:          removeNWBRequest.HealthcheckType,
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

func (newNWBRequest *NewBalanceInfo) convertDataForNWBService() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range newNWBRequest.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (removeNWBRequest *RemoveBalanceInfo) convertDataForNWBService() map[string]string {
	applicationServersMap := map[string]string{}
	for _, d := range removeNWBRequest.ApplicationServers {
		applicationServersMap[d.ServerIP] = d.ServerPort
	}
	return applicationServersMap
}

func (newNWBRequest *NewBalanceInfo) validateIncomeRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortNewBalanceInfoValidation, NewBalanceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	err := validate.Struct(newNWBRequest)
	if err != nil {
		return err
	}
	return nil
}

func customPortNewBalanceInfoValidation(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(NewBalanceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}

func customPortServerApplicationValidation(sl validator.StructLevel) {
	sA := sl.Current().Interface().(ServerApplication)
	port, err := strconv.Atoi(sA.ServerPort)
	if err != nil {
		sl.ReportError(sA.ServerPort, "serverPort", "ServerPort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(sA.ServerPort, "serverPort", "ServerPort", "port must gt=0 and lt=20000", "")
	}
}

func (removeNWBRequest *RemoveBalanceInfo) validateIncomeRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortNewBalanceInfoValidation, NewBalanceInfo{})
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

func errorsValidateToString(validateError error) string {
	var errorsString string
	for _, err := range validateError.(validator.ValidationErrors) {
		errorsString += fmt.Sprintf("In data %v got %v, can't validate in for rule %v %v\n",
			err.Field(),
			err.Value(),
			err.ActualTag(),
			err.Param())
	}
	return errorsString
}
