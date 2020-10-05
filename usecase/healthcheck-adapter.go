package usecase

import (
	"sync"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
)

// TODO: move HC to other service
// HCService ...
type HCService struct {
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
	Address             string              `json:"address"`
	IP                  string              `json:"ip"`
	Port                string              `json:"port"`
	IsUp                bool                `json:"isUp"`
	HCAddress           string              `json:"hcAddress"`
	HCApplicationServer HCApplicationServer `json:"hcApplicationServer"`
	ExampleBashCommands string              `json:"-"`
}

type HCApplicationServer struct {
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

func convertDomainServiceToHCService(domainServiceInfo *domain.ServiceInfo) *HCService {
	preparedApplicationServers := convertDomainApplicationServersToHCApplicationServers(domainServiceInfo.ApplicationServers)
	hcService := &HCService{}
	hcService.Address = domainServiceInfo.Address
	hcService.IP = domainServiceInfo.IP
	hcService.Port = domainServiceInfo.Port
	// hcService.IsUp
	hcService.BalanceType = domainServiceInfo.BalanceType
	hcService.RoutingType = domainServiceInfo.RoutingType
	hcService.Protocol = domainServiceInfo.Protocol
	hcService.AlivedAppServersForUp = domainServiceInfo.AlivedAppServersForUp
	hcService.HCType = domainServiceInfo.HCType
	hcService.HCRepeat = domainServiceInfo.HCRepeat
	hcService.HCTimeout = domainServiceInfo.HCTimeout
	hcService.HCNearFieldsMode = domainServiceInfo.HCNearFieldsMode
	hcService.HCUserDefinedData = domainServiceInfo.HCUserDefinedData
	hcService.HCRetriesForUP = domainServiceInfo.HCRetriesForUP
	hcService.HCRetriesForDown = domainServiceInfo.HCRetriesForDown
	hcService.ApplicationServers = preparedApplicationServers
	return hcService
}

func convertDomainApplicationServersToHCApplicationServers(domainApplicationServers []*domain.ApplicationServer) []*ApplicationServer {
	preparedApplicationServers := make([]*ApplicationServer, len(domainApplicationServers))
	for i, domainApplicationServer := range domainApplicationServers {
		preparedApplicationServer := &ApplicationServer{}
		preparedApplicationServer.Address = domainApplicationServer.Address
		preparedApplicationServer.IP = domainApplicationServer.IP
		preparedApplicationServer.Port = domainApplicationServer.Port
		// preparedApplicationServer.IsUp
		preparedApplicationServer.HCAddress = domainApplicationServer.HCAddress
		preparedApplicationServer.ExampleBashCommands = domainApplicationServer.ExampleBashCommands
		preparedApplicationServers[i] = preparedApplicationServer
	}
	return preparedApplicationServers
}

func convertHCServiceToDomainServiceInfo(hcService *HCService) *domain.ServiceInfo {
	preparedApplicationServers := convertHCApplicationServersToDomainApplicationServers(hcService.ApplicationServers)
	domainServiceInfo := &domain.ServiceInfo{}
	domainServiceInfo.Address = hcService.Address
	domainServiceInfo.IP = hcService.IP
	domainServiceInfo.Port = hcService.Port
	domainServiceInfo.IsUp = hcService.IsUp // unknown state = hcService.IsUp
	domainServiceInfo.BalanceType = hcService.BalanceType
	domainServiceInfo.RoutingType = hcService.RoutingType
	domainServiceInfo.Protocol = hcService.Protocol
	domainServiceInfo.AlivedAppServersForUp = hcService.AlivedAppServersForUp
	domainServiceInfo.HCType = hcService.HCType
	domainServiceInfo.HCRepeat = hcService.HCRepeat
	domainServiceInfo.HCTimeout = hcService.HCTimeout
	domainServiceInfo.HCNearFieldsMode = hcService.HCNearFieldsMode
	domainServiceInfo.HCUserDefinedData = hcService.HCUserDefinedData
	domainServiceInfo.HCRetriesForUP = hcService.HCRetriesForUP
	domainServiceInfo.HCRetriesForDown = hcService.HCRetriesForDown
	domainServiceInfo.ApplicationServers = preparedApplicationServers
	domainServiceInfo.HCStop = make(chan struct{}, 1)
	domainServiceInfo.HCStopped = make(chan struct{}, 1)
	return domainServiceInfo
}

func convertHCApplicationServersToDomainApplicationServers(restApplicationServers []*ApplicationServer) []*domain.ApplicationServer {
	preparedApplicationServers := make([]*domain.ApplicationServer, len(restApplicationServers))
	for i, restApplicationServer := range restApplicationServers {
		preparedApplicationServer := &domain.ApplicationServer{}
		preparedApplicationServer.Address = restApplicationServer.Address
		preparedApplicationServer.IP = restApplicationServer.IP
		preparedApplicationServer.Port = restApplicationServer.Port
		preparedApplicationServer.IsUp = restApplicationServer.IsUp
		preparedApplicationServer.HCAddress = restApplicationServer.HCAddress
		preparedApplicationServers[i] = preparedApplicationServer
	}
	return preparedApplicationServers
}
