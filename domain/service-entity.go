package domain

import "time"

// ServiceHealthcheck ...
type ServiceHealthcheck struct {
	Type    string        `json:"type" example:"http"`
	Timeout time.Duration `json:"timeout" example:"2s"`
}

// ServerHealthcheck ...
type ServerHealthcheck struct {
	HealthcheckAddress string `json:"address"` // ip+port, http address or some one else
}

// ApplicationServer ...
type ApplicationServer struct {
	ServerIP           string            `json:"serverIP"`
	ServerPort         string            `json:"serverPort"`
	State              bool              `json:"state"`
	IfcfgTunnelFile    string            `json:"ifcfgTunnelFile"` // full path to ifcfg file
	RouteTunnelFile    string            `json:"tunnelFile"`      // full path to route file
	SysctlConfFile     string            `json:"sysctlConf"`      // full path to sysctl conf file
	TunnelName         string            `json:"tunnelName"`
	ServerHealthcheck  ServerHealthcheck `json:"serverHealthcheck"`
	ServerBashCommands string            `json:"-"`
}

// ServiceInfo ...
type ServiceInfo struct {
	ServiceIP          string               `json:"serviceIP"`
	ServicePort        string               `json:"servicePort"`
	ApplicationServers []*ApplicationServer `json:"applicationServers"`
	Healthcheck        ServiceHealthcheck   `json:"serviceHealthcheck"`
	ExtraInfo          []string             `json:"extraInfo"`
	State              bool                 `json:"state"`
}

// ServiceWorker ...
type ServiceWorker interface {
	CreateService(*ServiceInfo, string) error
	RemoveService(*ServiceInfo, string) error
	AddApplicationServersForService(*ServiceInfo, string) error
	RemoveApplicationServersFromService(*ServiceInfo, string) error
	ReadCurrentConfig() ([]*ServiceInfo, error)
}
