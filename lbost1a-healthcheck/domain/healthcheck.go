package domain

import (
	"fmt"
	"sync"
	"time"
)

// Locker lock other commands for execute
type Locker struct {
	sync.Mutex
}

// HCService ...
type HCService struct {
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
	HCUserDefinedData     map[string]string      `json:"hcUserDefinedData,omitempty"`
	HCRetriesForUP        int                    `json:"hcRetriesForUP"`
	HCRetriesForDown      int                    `json:"hcRetriesForDown"`
	HCApplicationServers  []*HCApplicationServer `json:"ApplicationServers"`
	HCStop                chan struct{}          `json:"-"`
	HCStopped             chan struct{}          `json:"-"`
}

type HCApplicationServer struct {
	sync.RWMutex        `json:"-"`
	Address             string     `json:"address"`
	IP                  string     `json:"ip"`
	Port                string     `json:"port"`
	IsUp                bool       `json:"isUp"`
	HCAddress           string     `json:"hcAddress"`
	InternalHC          InternalHC `json:"hcApplicationServer"`
	ExampleBashCommands string     `json:"-"`
}

type InternalHC struct {
	HCType           string
	HCAddress        string
	HCTimeout        time.Duration
	RetriesForUP     []bool
	RetriesForDown   []bool
	LastIndexForUp   int
	LastIndexForDown int
	NearFieldsMode   bool
	UserDefinedData  map[string]string
}

// Release stringer interface for print/log data in []*ApplicationServer
func (hcApplicationServer *HCApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{Address:%s, IsUp:%v, HCAddress:%v, HCType:%v, HCTimeout:%v, RetriesForUP:%v, RetriesForDown:%v, LastIndexForUp:%v, LastIndexForDown:%v, NearFieldsMode:%v, UserDefinedData:%v}",
		hcApplicationServer.Address,
		hcApplicationServer.IsUp,
		hcApplicationServer.HCAddress,
		hcApplicationServer.InternalHC.HCType,
		hcApplicationServer.InternalHC.HCTimeout,
		hcApplicationServer.InternalHC.RetriesForUP,
		hcApplicationServer.InternalHC.RetriesForDown,
		hcApplicationServer.InternalHC.LastIndexForUp,
		hcApplicationServer.InternalHC.LastIndexForDown,
		hcApplicationServer.InternalHC.NearFieldsMode,
		hcApplicationServer.InternalHC.UserDefinedData,
	)
}
