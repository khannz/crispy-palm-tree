package domain

// ApplicationServer ...
type ApplicationServer struct {
	ServerIP           string `json:"serverIP"`
	ServerPort         string `json:"serverPort"`
	ServerBashCommands string `json:"-"`
}

// ServiceInfo ...
type ServiceInfo struct {
	ServiceIP          string              `json:"serviceIP"`
	ServicePort        string              `json:"servicePort"`
	ApplicationServers []ApplicationServer `json:"applicationServers"`
	HealthcheckType    string              `json:"healthcheckType"`
	ExtraInfo          []string            `json:"extraInfo"`
}

// ServiceWorker ...
type ServiceWorker interface {
	CreateService(ServiceInfo, string) error
	RemoveService(ServiceInfo, string) error
	AddApplicationServersFromService(ServiceInfo, string) error
	RemoveApplicationServersFromService(ServiceInfo, string) error
}
