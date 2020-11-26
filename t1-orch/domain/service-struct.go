package domain

import (
	"fmt"
	"sync"
	"time"
)

// ServiceInfo ...
type ServiceInfo struct {
	sync.RWMutex          `json:"-"`
	Address               string                        `json:"address"`
	IP                    string                        `json:"ip"`
	Port                  string                        `json:"port"`
	IsUp                  bool                          `json:"isUp"`
	BalanceType           string                        `json:"balanceType"`
	RoutingType           string                        `json:"routingType"`
	Protocol              string                        `json:"protocol"`
	AlivedAppServersForUp int                           `json:"alivedAppServersForUp"`
	HCType                string                        `json:"hcType"`
	HCRepeat              time.Duration                 `json:"hcRepeat"`
	HCTimeout             time.Duration                 `json:"hcTimeout"`
	HCNearFieldsMode      bool                          `json:"hcNearFieldsMode,omitempty"`
	HCUserDefinedData     map[string]string             `json:"hcUserDefinedData,omitempty"`
	HCRetriesForUP        int                           `json:"hcRetriesForUP"`
	HCRetriesForDown      int                           `json:"hcRetriesForDown"`
	ApplicationServers    map[string]*ApplicationServer `json:"ApplicationServers"`
	HCStop                chan struct{}                 `json:"-"`
	HCStopped             chan struct{}                 `json:"-"`
}

// ApplicationServer ...
type ApplicationServer struct {
	sync.RWMutex `json:"-"`
	Address      string     `json:"address"`
	IP           string     `json:"ip"`
	Port         string     `json:"port"`
	IsUp         bool       `json:"isUp"`
	HCAddress    string     `json:"hcAddress"`
	InternalHC   InternalHC `json:"hcApplicationServer"`
}

type InternalHC struct {
	HCType           string
	HCAddress        string
	HCTimeout        time.Duration
	RetriesForUP     []bool
	RetriesForDown   []bool
	LastIndexForUp   int
	LastIndexForDown int
	Mark             int
	NearFieldsMode   bool
	UserDefinedData  map[string]string
}

// Release stringer interface for print/log data in map[string]*ApplicationServer
func (applicationServer *ApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{Address:%s, IsUp:%v, HCAddress:%v, HCType:%v, HCTimeout:%v, RetriesForUP:%v, RetriesForDown:%v, LastIndexForUp:%v, LastIndexForDown:%v, NearFieldsMode:%v, UserDefinedData:%v}",
		applicationServer.Address,
		applicationServer.IsUp,
		applicationServer.HCAddress,
		applicationServer.InternalHC.HCType,
		applicationServer.InternalHC.HCTimeout,
		applicationServer.InternalHC.RetriesForUP,
		applicationServer.InternalHC.RetriesForDown,
		applicationServer.InternalHC.LastIndexForUp,
		applicationServer.InternalHC.LastIndexForDown,
		applicationServer.InternalHC.NearFieldsMode,
		applicationServer.InternalHC.UserDefinedData,
	)
}
