package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

// ServiceHealthcheck ...
type ServiceHealthcheck struct {
	Type                 string        `json:"type" validate:"required" example:"tcp"`
	Timeout              time.Duration `json:"timeout" validate:"required" example:"1000000000"`
	RepeatHealthcheck    time.Duration `json:"repeatHealthcheck" validate:"required" example:"3000000000"`
	PercentOfAlivedForUp int           `json:"percentOfAlivedForUp" validate:"required,gt=0,lte=100"`
}

// AdvancedHealthcheckParameters ...
type AdvancedHealthcheckParameters struct {
	NearFieldsMode  bool                   `json:"nearFieldsMode" validate:"required" example:"false"`
	UserDefinedData map[string]interface{} `json:"userDefinedData" validate:"required" example:"{'process_name': 'diskmon', 'status': 'Running'}"`
}

// ServerHealthcheck ...
type ServerHealthcheck struct {
	TypeOfCheck                   string                          `json:"typeOfCheck,omitempty" example:"http-advanced-json"`
	HealthcheckAddress            string                          `json:"healthcheckAddress,omitempty"`            // FIXME: need extra validate; ip+port, http address or some one else
	AdvancedHealthcheckParameters []AdvancedHealthcheckParameters `json:"advancedHealthcheckParameters,omitempty"` // TODO: example
}

// ServerApplication ...
type ServerApplication struct {
	ServerIP                    string            `json:"ip" validate:"required,ipv4" example:"1.1.1.1"`
	ServerPort                  string            `json:"port" validate:"required" example:"1111"`
	ServerHealthcheck           ServerHealthcheck `json:"serverHealthcheck,omitempty"`
	ServerСonfigurationCommands string            `json:"bashCommands,omitempty" swaggerignore:"true"`
}

// UniversalResponse ...
type UniversalResponse struct {
	ID                       string              `json:"id,omitempty"`
	ApplicationServers       []ServerApplication `json:"applicationServers,omitempty"`
	ServiceIP                string              `json:"serviceIP,omitempty"`
	ServicePort              string              `json:"servicePort,omitempty"`
	Healthcheck              ServiceHealthcheck  `json:"healthcheck,omitempty"`
	JobCompletedSuccessfully bool                `json:"jobCompletedSuccessfully"`
	ExtraInfo                string              `json:"extraInfo,omitempty"`
	BalanceType              string              `json:"balanceType,omitempty"`
	RoutingType              string              `json:"routingType,omitempty"`
}

// ServerApplicationWithStates ...
type ServerApplicationWithStates struct {
	ServerIP                    string            `json:"ip" validate:"required,ipv4" example:"1.1.1.1"`
	ServerPort                  string            `json:"port" validate:"required" example:"1111"`
	ServerHealthcheck           ServerHealthcheck `json:"serverHealthcheck,omitempty"`
	IsUp                        bool              `json:"serverIsUp"`
	ServerСonfigurationCommands string            `json:"bashCommands,omitempty" swaggerignore:"true"`
}

// UniversalResponseWithStates ...
type UniversalResponseWithStates struct {
	ID                       string                        `json:"id,omitempty"`
	ApplicationServers       []ServerApplicationWithStates `json:"applicationServers,omitempty"`
	ServiceIP                string                        `json:"serviceIP,omitempty"`
	ServicePort              string                        `json:"servicePort,omitempty"`
	Healthcheck              ServiceHealthcheck            `json:"healthcheck,omitempty"`
	JobCompletedSuccessfully bool                          `json:"jobCompletedSuccessfully"`
	ExtraInfo                string                        `json:"extraInfo,omitempty"`
	BalanceType              string                        `json:"balanceType,omitempty"`
	RoutingType              string                        `json:"routingType,omitempty"`
	IsUp                     bool                          `json:"serviceIsUp"`
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

func customServiceHealthcheckValidation(sl validator.StructLevel) {
	sHc := sl.Current().Interface().(ServiceHealthcheck)
	sA := sl.Current().Interface().([]ServerApplication)
	switch sHc.Type {
	case "tcp":
	case "http":
	case "icmp":
	case "http-advanced":
		if sA != nil && len(sA) > 0 {
			for i, applicationServer := range sA {
				switch applicationServer.ServerHealthcheck.TypeOfCheck {
				case "http-advanced-json":
					if sA[i].ServerHealthcheck.AdvancedHealthcheckParameters == nil || len(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters) == 0 {
						sl.ReportError(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters,
							"advancedHealthcheckParameters",
							"AdvancedHealthcheckParameters",
							"At healthcheck type 'http-advanced-json' http advanced parameters must be set",
							"")
					}
					for j, advancedHealthcheckParameters := range sA[i].ServerHealthcheck.AdvancedHealthcheckParameters {
						if len(advancedHealthcheckParameters.UserDefinedData) == 0 {
							sl.ReportError(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters[j].UserDefinedData,
								"userDefinedData",
								"UserDefinedData",
								"At healthcheck type 'http-advanced-json' at http advanced parameters user defined data must be set",
								"")
						}
					}
				default:
					sl.ReportError(sA[i].ServerHealthcheck.TypeOfCheck,
						"typeOfCheck",
						"TypeOfCheck",
						"Unsupported type of http advanced healthcheck",
						"")
				}
			}
		}
	default:
		sl.ReportError(sHc.Type, "type", "Type", "unsupported healthcheck type", "")
	}
	if sHc.Timeout >= sHc.RepeatHealthcheck {
		sl.ReportError(sHc.Timeout, "timeout", "Timeout", "timeout can't be more than repeat healthcheck", "")
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

func transformSliceToString(slice []string) string {
	var resultString string
	for _, el := range slice {
		resultString += "\n" + el
	}
	return resultString
}

func logNewRequest(typeOfRequest, uuid string, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Infof("got new %v request", typeOfRequest)
}

// read incoming data to buffer
func readIncomeBytes(req *http.Request) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	return buf.Bytes()
}

func unmarshallIncomeError(errS, uuid string, w http.ResponseWriter, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Errorf("can't unmarshal income request: %v", errS)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	rError := &UniversalResponse{
		ID:                       uuid,
		JobCompletedSuccessfully: false,
		ExtraInfo:                "can't unmarshal income request: " + errS,
	}
	err := json.NewEncoder(w).Encode(rError)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": uuid,
		}).Errorf("can't response by request: %v", err)
	}
}

func logChangeUUID(oldUUID, newUUID string, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newUUID,
	}).Infof("change job uuid from %v to %v", oldUUID, newUUID)
}

func validateIncomeError(errS, uuid string, w http.ResponseWriter, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Errorf("validate fail for income nwb request: %v", errS)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	rError := &UniversalResponse{
		ID:                       uuid,
		JobCompletedSuccessfully: false,
		ExtraInfo:                "fail when validate income request: " + errS,
	}
	err := json.NewEncoder(w).Encode(rError)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": uuid,
		}).Errorf("can't response by request: %v", err)
	}
}

func uscaseFail(typeOfrequest, errS, uuid string, w http.ResponseWriter, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Errorf("can't %v, got error: %v", typeOfrequest, errS)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	rError := &UniversalResponse{
		ID:                       uuid,
		JobCompletedSuccessfully: false,
		ExtraInfo:                "can't %v, got internal error: " + errS,
	}
	err := json.NewEncoder(w).Encode(rError)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": uuid,
		}).Errorf("can't response by %v request: %v", typeOfrequest, err)
	}
}

func logRequestIsDone(typeOfrequest, uuid string, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Infof("request %v done", typeOfrequest)
}

func writeUniversalResponse(ur UniversalResponse, typeOfrequest, uuid string, w http.ResponseWriter, logging *logrus.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ur); err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": uuid,
		}).Errorf("can't response by request: %v", err)
	}
}

func convertDomainHealthcheckToRest(dHC domain.ServiceHealthcheck) ServiceHealthcheck {
	return ServiceHealthcheck{
		Type:                 dHC.Type,
		Timeout:              dHC.Timeout,
		RepeatHealthcheck:    dHC.RepeatHealthcheck,
		PercentOfAlivedForUp: dHC.PercentOfAlivedForUp,
	}
}

func convertDomainApplicationServersToRest(dAS []*domain.ApplicationServer) []ServerApplication {
	sas := []ServerApplication{}
	for _, dSA := range dAS {
		arrayOfAdvancedHealthcheckParameters := []AdvancedHealthcheckParameters{}
		for _, aHP := range dSA.ServerHealthcheck.AdvancedHealthcheckParameters {
			advancedHealthcheckParameter := AdvancedHealthcheckParameters{
				NearFieldsMode:  aHP.NearFieldsMode,
				UserDefinedData: aHP.UserDefinedData,
			}
			arrayOfAdvancedHealthcheckParameters = append(arrayOfAdvancedHealthcheckParameters, advancedHealthcheckParameter)
		}

		svHCAdr := ServerHealthcheck{
			TypeOfCheck:                   dSA.ServerHealthcheck.TypeOfCheck,
			HealthcheckAddress:            dSA.ServerHealthcheck.HealthcheckAddress,
			AdvancedHealthcheckParameters: arrayOfAdvancedHealthcheckParameters,
		}
		sa := ServerApplication{
			ServerIP:                    dSA.ServerIP,
			ServerPort:                  dSA.ServerPort,
			ServerHealthcheck:           svHCAdr,
			ServerСonfigurationCommands: dSA.ServerСonfigurationCommands,
		}
		sas = append(sas, sa)
	}
	return sas
}

func convertDomainServiceInfoToRestUniversalResponse(serviceInfo *domain.ServiceInfo, isOk bool) UniversalResponse {
	return UniversalResponse{
		ApplicationServers:       convertDomainApplicationServersToRest(serviceInfo.ApplicationServers),
		ServiceIP:                serviceInfo.ServiceIP,
		ServicePort:              serviceInfo.ServicePort,
		Healthcheck:              convertDomainHealthcheckToRest(serviceInfo.Healthcheck),
		JobCompletedSuccessfully: isOk,
		ExtraInfo:                transformSliceToString(serviceInfo.ExtraInfo),
		BalanceType:              serviceInfo.BalanceType,
		RoutingType:              serviceInfo.RoutingType,
	}
}

func convertDomainServicesInfoToRestUniversalResponse(servicesInfo []*domain.ServiceInfo, isOk bool) []UniversalResponse {
	urs := []UniversalResponse{}
	for _, serviceInfo := range servicesInfo {
		urs = append(urs, convertDomainServiceInfoToRestUniversalResponse(serviceInfo, isOk))
	}
	return urs
}

func convertDomainServiceInfoToRestUniversalResponseWithStates(serviceInfo *domain.ServiceInfo, isOk bool) UniversalResponseWithStates {
	return UniversalResponseWithStates{
		ApplicationServers:       convertDomainApplicationServersToRestWithState(serviceInfo.ApplicationServers),
		ServiceIP:                serviceInfo.ServiceIP,
		ServicePort:              serviceInfo.ServicePort,
		Healthcheck:              convertDomainHealthcheckToRest(serviceInfo.Healthcheck),
		JobCompletedSuccessfully: isOk,
		ExtraInfo:                transformSliceToString(serviceInfo.ExtraInfo),
		BalanceType:              serviceInfo.BalanceType,
		RoutingType:              serviceInfo.RoutingType,
		IsUp:                     serviceInfo.IsUp,
	}
}

func convertDomainApplicationServersToRestWithState(dAS []*domain.ApplicationServer) []ServerApplicationWithStates {
	sas := []ServerApplicationWithStates{}
	for _, dSA := range dAS {
		arrayOfAdvancedHealthcheckParameters := []AdvancedHealthcheckParameters{}
		for _, aHP := range dSA.ServerHealthcheck.AdvancedHealthcheckParameters {
			advancedHealthcheckParameter := AdvancedHealthcheckParameters{
				NearFieldsMode:  aHP.NearFieldsMode,
				UserDefinedData: aHP.UserDefinedData,
			}
			arrayOfAdvancedHealthcheckParameters = append(arrayOfAdvancedHealthcheckParameters, advancedHealthcheckParameter)
		}

		svHCAdr := ServerHealthcheck{
			TypeOfCheck:                   dSA.ServerHealthcheck.TypeOfCheck,
			HealthcheckAddress:            dSA.ServerHealthcheck.HealthcheckAddress,
			AdvancedHealthcheckParameters: arrayOfAdvancedHealthcheckParameters,
		}
		sa := ServerApplicationWithStates{
			ServerIP:          dSA.ServerIP,
			ServerPort:        dSA.ServerPort,
			ServerHealthcheck: svHCAdr,
			IsUp:              dSA.IsUp,
		}
		sas = append(sas, sa)
	}
	return sas
}

func convertDomainServicesInfoToRestUniversalResponseWithState(servicesInfo []*domain.ServiceInfo, isOk bool) []UniversalResponseWithStates {
	urs := []UniversalResponseWithStates{}
	for _, serviceInfo := range servicesInfo {
		urs = append(urs, convertDomainServiceInfoToRestUniversalResponseWithStates(serviceInfo, isOk))
	}
	return urs
}

func validateServiceBalanceType(balanceType string) error {
	switch balanceType { // maybe range by array is better?
	case "rr":
	case "wrr":
	case "lc":
	case "wlc":
	case "lblc":
	case "sh":
	case "mh":
	case "dh":
	case "fo":
	case "ovf":
	case "lblcr":
	case "sed":
	case "nq":
	default:
		return fmt.Errorf("unknown balance type for service: %v; supported types: rr|wrr|lc|wlc|lblc|sh|mh|dh|fo|ovf|lblcr|sed|nq0", balanceType)
	}
	return nil
}

func validateServiceRoutingType(routingType string) error {
	switch routingType { // maybe range by array is better?
	case "masquerading":
	case "tunneling":
	default:
		return fmt.Errorf("unknown routing type for service: %v; supported types: masquerading|tunneling", routingType)
	}
	return nil
}
