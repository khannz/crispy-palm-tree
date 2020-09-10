package application

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

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
	// sA := sl.Current().Interface().([]ServerApplication) // FIXME: broken
	switch sHc.Type {
	case "tcp":
	case "http":
	case "icmp":
	case "http-advanced":
		// if sA != nil && len(sA) > 0 {
		// 	for i, applicationServer := range sA {
		// 		switch applicationServer.ServerHealthcheck.TypeOfCheck {
		// 		case "http-advanced-json":
		// 			if sA[i].ServerHealthcheck.AdvancedHealthcheckParameters == nil || len(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters) == 0 {
		// 				sl.ReportError(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters,
		// 					"advancedHealthcheckParameters",
		// 					"AdvancedHealthcheckParameters",
		// 					"At healthcheck type 'http-advanced-json' http advanced parameters must be set",
		// 					"")
		// 			}
		// 			for j, advancedHealthcheckParameters := range sA[i].ServerHealthcheck.AdvancedHealthcheckParameters {
		// 				if len(advancedHealthcheckParameters.UserDefinedData) == 0 {
		// 					sl.ReportError(sA[i].ServerHealthcheck.AdvancedHealthcheckParameters[j].UserDefinedData,
		// 						"userDefinedData",
		// 						"UserDefinedData",
		// 						"At healthcheck type 'http-advanced-json' at http advanced parameters user defined data must be set",
		// 						"")
		// 				}
		// 			}
		// 		default:
		// 			sl.ReportError(sA[i].ServerHealthcheck.TypeOfCheck,
		// 				"typeOfCheck",
		// 				"TypeOfCheck",
		// 				"Unsupported type of http advanced healthcheck",
		// 				"")
		// 		}
		// 	}
		// }
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

func unmarshallIncomeError(errS, uuid string, ginContext *gin.Context, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Errorf("can't unmarshal income request: %v", errS)

	rError := UniversalResponse{
		ID:                       uuid,
		JobCompletedSuccessfully: false,
		ExtraInfo:                "can't unmarshal income request: " + errS,
	}

	ginContext.JSON(http.StatusBadRequest, rError)
}

func logChangeUUID(oldUUID, newUUID string, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": newUUID,
	}).Infof("change job uuid from %v to %v", oldUUID, newUUID)
}

func validateIncomeError(errS, uuid string, ginContext *gin.Context, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Errorf("validate fail for income nwb request: %v", errS)

	rError := &UniversalResponse{
		ID:                       uuid,
		JobCompletedSuccessfully: false,
		ExtraInfo:                "fail when validate income request: " + errS,
	}
	ginContext.JSON(http.StatusBadRequest, rError)
}

func uscaseFail(typeOfrequest, errS, uuid string, ginContext *gin.Context, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Errorf("can't %v, got error: %v", typeOfrequest, errS)

	rError := &UniversalResponse{
		ID:                       uuid,
		JobCompletedSuccessfully: false,
		ExtraInfo:                "can't %v, got internal error: " + errS,
	}
	ginContext.JSON(http.StatusInternalServerError, rError)
}

func logRequestIsDone(typeOfrequest, uuid string, logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": uuid,
	}).Infof("request %v done", typeOfrequest)
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
		Protocol:                 serviceInfo.Protocol,
	}
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
		Protocol:                 serviceInfo.Protocol,
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

func validateServiceProtocol(protocol string) error {
	switch protocol { // maybe range by array is better?
	case "tcp":
	case "udp":
	default:
		return fmt.Errorf("unknown protocol for service: %v; supported types: tcp|udp", protocol)
	}
	return nil
}
