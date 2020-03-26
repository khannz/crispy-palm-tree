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
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

const restAPIlogName = "restAPI"

// @title Netcon agent swagger
// @version 1.0.1
// @description create/delete nlb service.
// @Tags New nlb service
// @tag.name Link for docs
// @tag.docs.url http://kb.sdn.sbrf.ru/display/SDN/*
// @tag.docs.description Docs at confluence
// @contact.name Ivan Tikhonov
// @contact.email sdn@sberbank.ru

type nwbResponse struct {
	State string `json:"state"`
}

// NewBalanceInfo ...
type NewBalanceInfo struct {
	JobUUID         string         `json:"jobUUID" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP       string         `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort     int            `json:"servicePort" validate:"min=1,max=20000" example:"1111"`
	RealServersData map[string]int `json:"realServersData" validate:"gt=0,dive,keys,ipv4,endkeys,min=1,max=20000"`
}

// RemoveBalanceInfo ...
type RemoveBalanceInfo struct {
	JobUUID         string         `json:"jobUUID" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP       string         `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort     int            `json:"servicePort" validate:"min=1,max=20000" example:"1111"`
	RealServersData map[string]int `json:"realServersData" validate:"gt=0,dive,keys,ipv4,endkeys,min=1,max=20000"`
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
	restAPI.router.PathPrefix("/swagger-ui.html/").Handler(httpSwagger.WrapHandler)

	err := restAPI.server.ListenAndServe()
	if err != nil {
		restAPI.balancerFacade.Logging.Infof("rest api down: %v", err)
	}
}

// newNWBRequest godoc
// @tags Network balance services
// @Summary Create nlb service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.NewBalanceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.nwbResponse "If all okay"
// @Failure 400 {object} application.nwbResponse "Bad request"
// @Failure 500 {object} application.nwbResponse "Internal error"
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
		rError := &nwbResponse{State: "can't unmarshal income nwb request"}
		json.NewEncoder(w).Encode(rError)
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newNWBRequestUUID,
	}).Infof("change job uuid from %v to %v", newNWBRequestUUID, newNWBRequest.JobUUID)
	newNWBRequestUUID = newNWBRequest.JobUUID

	validateError := newNWBRequest.validateIncomeRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newNWBRequestUUID,
		}).Errorf("validate fail for income nwb request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &nwbResponse{State: stringValidateError}
		json.NewEncoder(w).Encode(rError)
		return
	}
	servicePort, realServersData := newNWBRequest.convertDataForNWBService()
	err = restAPI.balancerFacade.NewNWBService(newNWBRequest.ServiceIP,
		servicePort,
		realServersData,
		newNWBRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": newNWBRequestUUID,
		}).Errorf("can't create new nwb, got error: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &nwbResponse{State: "can't create new nwb, got internal error: " + err.Error()}
		json.NewEncoder(w).Encode(rError)
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newNWBRequestUUID,
	}).Info("new nwb created")

	nwbCreated := nwbResponse{State: "new nwb created"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nwbCreated)
}

// removeNWBRequest godoc
// @tags Network balance services
// @Summary Remove nlb service
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.RemoveBalanceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.nwbResponse "If all okay"
// @Failure 400 {object} application.nwbResponse "Bad request"
// @Failure 500 {object} application.nwbResponse "Internal error"
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
		rError := &nwbResponse{State: "can't unmarshal income remove nwb request"}
		json.NewEncoder(w).Encode(rError)
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeNWBRequestUUID,
	}).Infof("change job uuid from %v to %v", removeNWBRequestUUID, removeNWBRequest.JobUUID)
	removeNWBRequestUUID = removeNWBRequest.JobUUID

	validateError := removeNWBRequest.validateIncomeRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeNWBRequestUUID,
		}).Errorf("validate fail for income remove nwb request: %v", stringValidateError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		rError := &nwbResponse{State: stringValidateError}
		json.NewEncoder(w).Encode(rError)
		return
	}
	servicePort, realServersData := removeNWBRequest.convertDataForNWBService()
	err = restAPI.balancerFacade.RemoveNWBService(removeNWBRequest.ServiceIP,
		servicePort,
		realServersData,
		removeNWBRequestUUID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": removeNWBRequestUUID,
		}).Errorf("can't remove old nwb, got error: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		rError := &nwbResponse{State: "can't remove old nwb, got internal error: " + err.Error()}
		json.NewEncoder(w).Encode(rError)
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": removeNWBRequestUUID,
	}).Info("nwb removed")

	nwbCreated := nwbResponse{State: "nwb removed"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nwbCreated)
}

func (newNWBRequest *NewBalanceInfo) convertDataForNWBService() (string, map[string]string) {
	servicePort := strconv.Itoa(newNWBRequest.ServicePort)
	realServiceData := map[string]string{}
	for k, v := range newNWBRequest.RealServersData {
		realServiceData[k] = strconv.Itoa(v)
	}
	return servicePort, realServiceData
}

func (removeNWBRequest *RemoveBalanceInfo) convertDataForNWBService() (string, map[string]string) {
	servicePort := strconv.Itoa(removeNWBRequest.ServicePort)
	realServiceData := map[string]string{}
	for k, v := range removeNWBRequest.RealServersData {
		realServiceData[k] = strconv.Itoa(v)
	}
	return servicePort, realServiceData
}

func (newNWBRequest *NewBalanceInfo) validateIncomeRequest() error {
	validate := validator.New()
	err := validate.Struct(newNWBRequest)
	if err != nil {
		return err

	}
	return nil
}

func (removeNWBRequest *RemoveBalanceInfo) validateIncomeRequest() error {
	validate := validator.New()
	err := validate.Struct(removeNWBRequest)
	if err != nil {
		return err

	}
	return nil
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
