package application

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/khannz/crispy-palm-tree/domain"
)

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

func deepValidateServiceInfo(serviceIP, servicePort string, applicationServers map[string]string) error {
	for appSrvIP, appSrvPort := range applicationServers {
		if serviceIP == appSrvIP &&
			servicePort == appSrvPort {
			return fmt.Errorf("service %v:%v equal application service %v:%v",
				serviceIP,
				servicePort,
				appSrvIP,
				appSrvPort)
		}
	}
	return nil
}

func transformDomainServiceInfoToResponseData(serviceInfo domain.ServiceInfo, isOk bool) UniversalResponse {
	return UniversalResponse{
		ApplicationServers:       transformDomainApplicationServersToRestApplicationServers(serviceInfo.ApplicationServers),
		ServiceIP:                serviceInfo.ServiceIP,
		ServicePort:              serviceInfo.ServicePort,
		HealthcheckType:          serviceInfo.HealthcheckType,
		JobCompletedSuccessfully: isOk,
		ExtraInfo:                transformSliceToString(serviceInfo.ExtraInfo),
	}
}

func transformDomainServicesInfoToResponseData(nwbServices []domain.ServiceInfo, isOk bool) []UniversalResponse {
	UniversalResponses := []UniversalResponse{}
	for _, nwbService := range nwbServices {
		UniversalResponses = append(UniversalResponses, transformDomainServiceInfoToResponseData(nwbService, isOk))
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
