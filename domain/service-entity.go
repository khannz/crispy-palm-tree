package domain

import (
	"sync"
	"time"
)

// ServiceHealthcheck ...
type ServiceHealthcheck struct {
	StopChecks           chan struct{} `json:"-"` // when need to stop checks
	PercentOfAlivedForUp int           `json:"percentOfAlivedForUp"`
	Type                 string        `json:"type" example:"http"`
	Timeout              time.Duration `json:"timeout" example:"1000000000"`
	RepeatHealthcheck    time.Duration `json:"repeatHealthcheck" example:"3000000000"`
}

// ServerHealthcheck ...
type ServerHealthcheck struct {
	HealthcheckAddress string `json:"address"` // ip+port, http address or some one else
}

// ApplicationServer ...
type ApplicationServer struct {
	ServerIP                    string            `json:"serverIP"`
	ServerPort                  string            `json:"serverPort"`
	IsUp                        bool              `json:"serverIsUp"`
	ServerHealthcheck           ServerHealthcheck `json:"serverHealthcheck"`
	ServerСonfigurationCommands string            `json:"-"`
}

// ServiceInfo ...
type ServiceInfo struct {
	sync.Mutex
	ServiceIP          string               `json:"serviceIP"`
	ServicePort        string               `json:"servicePort"`
	ApplicationServers []*ApplicationServer `json:"applicationServers"`
	Healthcheck        ServiceHealthcheck   `json:"serviceHealthcheck"`
	ExtraInfo          []string             `json:"extraInfo"`
	IsUp               bool                 `json:"serviceIsUp"`
	BalanceType        string               `json:"balanceType"`
}

// CommandGenerator ...
type CommandGenerator interface {
	GenerateCommandsForApplicationServers(*ServiceInfo, string) error
}
