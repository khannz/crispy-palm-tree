package domain

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ServiceInfo ...
type ServiceInfo struct {
	sync.RWMutex             `json:"-"`
	Address                  string                   `json:"address"`
	IP                       string                   `json:"ip"`
	Port                     string                   `json:"port"`
	IsUp                     bool                     `json:"isUp"`
	BalanceType              string                   `json:"balanceType"`
	RoutingType              string                   `json:"routingType"`
	Protocol                 string                   `json:"protocol"`
	Quorum                   int                      `json:"quorum"`
	HealthcheckType          string                   `json:"healthcheckType"`
	HelloTimer               time.Duration            `json:"helloTimer"`
	ResponseTimer            time.Duration            `json:"responseTimer"`
	HCNearFieldsMode         bool                     `json:"hcNearFieldsMode,omitempty"`
	HCUserDefinedData        map[string]string        `json:"hcUserDefinedData,omitempty"`
	AliveThreshold           int                      `json:"aliveThreshold"`
	DeadThreshold            int                      `json:"deadThreshold"`
	ApplicationServers       ApplicationServers       `json:"ApplicationServers"`
	Uri                      string                   `json:"uri"`                // only for http(s) hc types
	ValidResponseCodes       []int64                  `json:"validResponseCodes"` // only for http(s) hc types
	FailedApplicationServers FailedApplicationServers `json:"-"`
	HCStop                   chan struct{}            `json:"-"`
	HCStopped                chan struct{}            `json:"-"`
}

//ApplicationServers alias
type ApplicationServers = map[string]*ApplicationServer

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
	return fmt.Sprintf("Address: %v, IP: %v, Port: %v, IsUp: %v, BalanceType: %v, RoutingType: %v, Protocol: %v, Quorum: %v, HealthcheckType: %v, HelloTimer: %v, ResponseTimer: %v, HCNearFieldsMode: %v, HCUserDefinedData: %v, AliveThreshold: %v, DeadThreshold: %v, ApplicationServers: %v, Uri: %v, ValidResponseCodes: %v",
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
		serviceInfo.ApplicationServers,
		serviceInfo.Uri,
		serviceInfo.ValidResponseCodes)
}

type FailedApplicationServers struct {
	count int32
}

func (failedApplicationServers *FailedApplicationServers) Add(n int) {
	atomic.AddInt32(&failedApplicationServers.count, int32(n))
}

func (failedApplicationServers *FailedApplicationServers) Count() int {
	return int(atomic.LoadInt32(&failedApplicationServers.count))
}

func (failedApplicationServers *FailedApplicationServers) Set2Zero() {
	atomic.StoreInt32(&failedApplicationServers.count, 0)
}
