package domain

// ApplicationServer ...
type ApplicationServer struct {
	ServerIP           string `json:"serverIP"`
	ServerPort         string `json:"serverPort"`
	State              bool   `json:"state"`
	IfcfgTunnelFile    string `json:"ifcfgTunnelFile"` // full path to ifcfg file
	RouteTunnelFile    string `json:"tunnelFile"`      // full path to route file
	SysctlConfFile     string `json:"sysctlConf"`      // full path to sysctl conf file
	TunnelName         string `json:"tunnelName"`
	ServerBashCommands string `json:"-"`
}

// ServiceInfo ...
type ServiceInfo struct {
	ServiceIP          string              `json:"serviceIP"`
	ServicePort        string              `json:"servicePort"`
	ApplicationServers []ApplicationServer `json:"applicationServers"`
	HealthcheckType    string              `json:"healthcheckType"` // TODO: must be struct
	ExtraInfo          []string            `json:"extraInfo"`
	State              bool                `json:"state"`
}

// ServiceWorker ...
type ServiceWorker interface {
	CreateService(*ServiceInfo, string) error
	RemoveService(*ServiceInfo, string) error
	AddApplicationServersFromService(*ServiceInfo, string) error
	RemoveApplicationServersFromService(*ServiceInfo, string) error
}
