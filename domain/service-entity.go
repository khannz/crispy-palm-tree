package domain

// ApplicationServer ...
type ApplicationServer struct {
	ServerIP           string
	ServerPort         string
	ServerBashCommands string
}

// ServiceInfo ...
type ServiceInfo struct {
	ServiceIP          string
	ServicePort        string
	ApplicationServers []ApplicationServer
	HealthcheckType    string
	ExtraInfo          []string
}

// ServiceWorker ...
type ServiceWorker interface {
	CreateService(string, string, map[string]string, string) error // validate service and application servers (ip+port) not existing
	// RemoveService()
}
