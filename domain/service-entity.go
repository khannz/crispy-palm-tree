package domain

import (
	"fmt"
	"sync"
	"time"
)

// ServiceInfo ...
type ServiceInfo struct {
	sync.RWMutex          `json:"-"`
	Address               string               `json:"address"`
	IP                    string               `json:"ip"`
	Port                  string               `json:"port"`
	IsUp                  bool                 `json:"isUp"`
	BalanceType           string               `json:"balanceType"`
	RoutingType           string               `json:"routingType"`
	Protocol              string               `json:"protocol"`
	AlivedAppServersForUp int                  `json:"alivedAppServersForUp"`
	HCType                string               `json:"hcType"`
	HCRepeat              time.Duration        `json:"hcRepeat"`
	HCTimeout             time.Duration        `json:"hcTimeout"`
	HCNearFieldsMode      bool                 `json:"hcNearFieldsMode,omitempty"`
	HCUserDefinedData     map[string]string    `json:"hcUserDefinedData,omitempty"`
	HCRetriesForUP        int                  `json:"hcRetriesForUP"`
	HCRetriesForDown      int                  `json:"hcRetriesForDown"`
	ApplicationServers    []*ApplicationServer `json:"ApplicationServers"`
	HCStop                chan struct{}        `json:"-"`
	HCStopped             chan struct{}        `json:"-"`
}

type ApplicationServer struct {
	sync.RWMutex        `json:"-"`
	Address             string `json:"address"`
	IP                  string `json:"ip"`
	Port                string `json:"port"`
	IsUp                bool   `json:"isUp"`
	HCAddress           string `json:"hcAddress"`
	ExampleBashCommands string `json:"-"`
}

// CommandGenerator ...
type CommandGenerator interface {
	GenerateCommandsForApplicationServers(*ServiceInfo, string) error
}

// Release stringer interface for print/log data in []*ApplicationServer
func (applicationServer *ApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{Address:%v, IsUp:%v, HCAddress:%v}",
		applicationServer.Address,
		applicationServer.IsUp,
		applicationServer.HCAddress)
}
