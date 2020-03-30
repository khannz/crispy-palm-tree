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
