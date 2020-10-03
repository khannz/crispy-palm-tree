package domain

import (
	"fmt"
	"sync"
	"time"
)

// ServiceInfo ...
type ServiceInfo struct {
	sync.RWMutex          `json:"-"`
	Address               string                 `json:"address"`
	IP                    string                 `json:"ip"`
	Port                  string                 `json:"port"`
	IsUp                  bool                   `json:"isUp"`
	BalanceType           string                 `json:"balanceType"`
	RoutingType           string                 `json:"routingType"`
	Protocol              string                 `json:"protocol"`
	AlivedAppServersForUp int                    `json:"alivedAppServersForUp"`
	HCType                string                 `json:"hcType"`
	HCRepeat              time.Duration          `json:"hcRepeat"`
	HCTimeout             time.Duration          `json:"hcTimeout"`
	HCNearFieldsMode      bool                   `json:"hcNearFieldsMode,omitempty"`
	HCUserDefinedData     map[string]interface{} `json:"hcUserDefinedData,omitempty"`
	HCRetriesForUP        int                    `json:"hcRetriesForUP"`
	HCRetriesForDown      int                    `json:"hcRetriesForDown"`
	ApplicationServers    []*ApplicationServer   `json:"ApplicationServers"`
	HCStop                chan struct{}          `json:"-"`
	HCStopped             chan struct{}          `json:"-"`
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

// // ServiceInfo ...
// type ServiceInfo struct {
// 	sync.Mutex
// 	ServiceIP          string               `json:"serviceIP"`
// 	ServicePort        string               `json:"servicePort"`
// 	ApplicationServers []*ApplicationServer `json:"applicationServers"`
// 	Healthcheck        ServiceHealthcheck   `json:"serviceHealthcheck"`
// 	ExtraInfo          []string             `json:"extraInfo"`
// 	IsUp               bool                 `json:"serviceIsUp"`
// 	BalanceType        string               `json:"balanceType"`
// 	RoutingType        string               `json:"routingType"`
// 	Protocol           string               `json:"protocol"`
// }

// // ApplicationServer ...
// type ApplicationServer struct {
// 	ServerIP                    string            `json:"serverIP"`
// 	ServerPort                  string            `json:"serverPort"`
// 	IsUp                        bool              `json:"serverIsUp"`
// 	ServerHealthcheck           ServerHealthcheck `json:"serverHealthcheck"`
// 	ServerСonfigurationCommands string            `json:"-"`
// }

// // ServiceHealthcheck ...
// type ServiceHealthcheck struct {
// 	StopChecks                      chan struct{} `json:"-"` // when we need to say stop checks
// 	ChecksStoped                    chan struct{} `json:"-"` // when checks stoped
// 	PercentOfAlivedForUp            int           `json:"percentOfAlivedForUp"`
// 	Type                            string        `json:"type" example:"http"`
// 	Timeout                         time.Duration `json:"timeout" example:"1000000000"`
// 	RepeatHealthcheck               time.Duration `json:"repeatHealthcheck" example:"3000000000"`
// 	RetriesForUpApplicationServer   int           `json:"retriesForUpApplicationServer"`
// 	RetriesForDownApplicationServer int           `json:"retriesForDownApplicationServer"`
// }

// // AdvancedHealthcheckParameters ...
// type AdvancedHealthcheckParameters struct {
// 	NearFieldsMode  bool                   `json:"nearFieldsMode"`
// 	UserDefinedData map[string]interface{} `json:"userDefinedData"`
// }

// // ServerHealthcheck ...
// type ServerHealthcheck struct {
// 	TypeOfCheck                   string                          `json:"typeOfCheck,omitempty"`
// 	HealthcheckAddress            string                          `json:"address"` // ip+port, http address or some one else
// 	AdvancedHealthcheckParameters []AdvancedHealthcheckParameters `json:"advancedHealthcheckParameters,omitempty"`
// 	RetriesCounterForUp           []bool                          `json:"-"` // internal for retries
// 	RetriesCounterForDown         []bool                          `json:"-"` // internal for retries
// 	LastIndexForUp                int                             `json:"-"` // internal for retries
// 	LastIndexForDown              int                             `json:"-"` // internal for retries
// }

// CommandGenerator ...
type CommandGenerator interface {
	GenerateCommandsForApplicationServers(*ServiceInfo, string) error
}

// Release stringer interface for print/log data in []*ApplicationServer
func (applicationServer *ApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{Address:%s, IsUp:%v, HCAddress:%v}",
		applicationServer.Address,
		applicationServer.IsUp,
		applicationServer.HCAddress)
}
