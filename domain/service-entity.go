package domain

import (
	"sync"
	"time"
)

// ServiceHealthcheck ...
type ServiceHealthcheck struct {
	StopChecks        chan struct{} `json:"-"` // when need to stop checks
	Type              string        `json:"type" example:"http"`
	Timeout           time.Duration `json:"timeout" example:"1000000000"`
	RepeatHealthcheck time.Duration `json:"repeatHealthcheck" example:"3000000000"`
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
	sync.Mutex
	ServiceIP          string               `json:"serviceIP"`
	ServicePort        string               `json:"servicePort"`
	ApplicationServers []*ApplicationServer `json:"applicationServers"`
	Healthcheck        ServiceHealthcheck   `json:"serviceHealthcheck"`
	ExtraInfo          []string             `json:"extraInfo"`
	State              bool                 `json:"state"`
}
