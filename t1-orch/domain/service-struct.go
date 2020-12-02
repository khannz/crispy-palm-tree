package domain

import (
	"fmt"
	"sync"
	"time"
)

// ServiceInfo ...
type ServiceInfo struct {
	sync.RWMutex             `json:"-"`
	Address                  string                        `json:"address"`
	IP                       string                        `json:"ip"`
	Port                     string                        `json:"port"`
	IsUp                     bool                          `json:"isUp"`
	BalanceType              string                        `json:"balanceType"`
	RoutingType              string                        `json:"routingType"`
	Protocol                 string                        `json:"protocol"`
	Quorum                   int                           `json:"quorum"`
	HealthcheckType          string                        `json:"healthcheckType"`
	HelloTimer               time.Duration                 `json:"helloTimer"`
	ResponseTimer            time.Duration                 `json:"responseTimer"`
	HCNearFieldsMode         bool                          `json:"hcNearFieldsMode,omitempty"`
	HCUserDefinedData        map[string]string             `json:"hcUserDefinedData,omitempty"`
	AliveThreshold           int                           `json:"aliveThreshold"`
	DeadThreshold            int                           `json:"deadThreshold"`
	ApplicationServers       map[string]*ApplicationServer `json:"ApplicationServers"`
	FailedApplicationServers *FailedApplicationServers     `json:"-"`
	HCStop                   chan struct{}                 `json:"-"`
	HCStopped                chan struct{}                 `json:"-"`
}

// ApplicationServer ...
type ApplicationServer struct {
	sync.RWMutex       `json:"-"`
	Address            string     `json:"address"`
	IP                 string     `json:"ip"`
	Port               string     `json:"port"`
	IsUp               bool       `json:"isUp"`
	HealthcheckAddress string     `json:"healthcheckAddress"`
	InternalHC         InternalHC `json:"hcApplicationServer"`
}

type InternalHC struct {
	HealthcheckType    string
	HealthcheckAddress string
	ResponseTimer      time.Duration
	AliveThreshold     []bool
	DeadThreshold      []bool
	LastIndexForAlive  int
	LastIndexForDead   int
	Mark               int
	NearFieldsMode     bool
	UserDefinedData    map[string]string
}

// Release stringer interface for print/log data in map[string]*ApplicationServer
func (applicationServer *ApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{Address:%s, IsUp:%v, HealthcheckAddress:%v, HealthcheckType:%v, ResponseTimer:%v, AliveThreshold:%v, DeadThreshold:%v, LastIndexForAlive:%v, LastIndexForDead:%v, NearFieldsMode:%v, UserDefinedData:%v}",
		applicationServer.Address,
		applicationServer.IsUp,
		applicationServer.HealthcheckAddress,
		applicationServer.InternalHC.HealthcheckType,
		applicationServer.InternalHC.ResponseTimer,
		applicationServer.InternalHC.AliveThreshold,
		applicationServer.InternalHC.DeadThreshold,
		applicationServer.InternalHC.LastIndexForAlive,
		applicationServer.InternalHC.LastIndexForDead,
		applicationServer.InternalHC.NearFieldsMode,
		applicationServer.InternalHC.UserDefinedData,
	)
}

// Release stringer interface for print/log data in []*ServiceInfo
func (serviceInfo *ServiceInfo) String() string {
	return fmt.Sprintf("Address: %v, IP: %v, Port: %v, IsUp: %v, BalanceType: %v, RoutingType: %v, Protocol: %v, Quorum: %v, HealthcheckType: %v, HelloTimer: %v, ResponseTimer: %v, HCNearFieldsMode: %v, HCUserDefinedData: %v, AliveThreshold: %v, DeadThreshold: %v, ApplicationServers: %v",
		serviceInfo.Address,
		serviceInfo.IP,
		serviceInfo.Port,
		serviceInfo.IsUp,
		serviceInfo.BalanceType,
		serviceInfo.RoutingType,
		serviceInfo.Protocol,
		serviceInfo.Quorum,
		serviceInfo.HealthcheckType,
		serviceInfo.HelloTimer,
		serviceInfo.ResponseTimer,
		serviceInfo.HCNearFieldsMode,
		serviceInfo.HCUserDefinedData,
		serviceInfo.AliveThreshold,
		serviceInfo.DeadThreshold,
		serviceInfo.ApplicationServers)
}

type FailedApplicationServers struct {
	sync.Mutex
	Wg    *sync.WaitGroup
	Count int
}

func NewFailedApplicationServers() *FailedApplicationServers {
	return &FailedApplicationServers{
		Wg:    new(sync.WaitGroup),
		Count: 0,
	}
}

func (failedApplicationServers *FailedApplicationServers) SetFailedApplicationServersToZero() {
	failedApplicationServers.Lock()
	defer failedApplicationServers.Unlock()
	failedApplicationServers.Count = 0
}
