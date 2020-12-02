package application

import (
	"math/big"
	"net"
	"strconv"
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

func convertBalancingServicesTransportArrayToDomainModel(servicesTransport []*ServiceTransport) ([]*domain.ServiceInfo, error) {
	servicesInfo := make([]*domain.ServiceInfo, 0, len(servicesTransport))
	for _, serviceTransport := range servicesTransport {

		quorum, err := strconv.Atoi(serviceTransport.Quorum)
		if err != nil {
			return nil, err
		}

		helloTimer, err := time.ParseDuration(serviceTransport.HelloTimer)
		if err != nil {
			return nil, err
		}

		responseTimer, err := time.ParseDuration(serviceTransport.ResponseTimer)
		if err != nil {
			return nil, err
		}

		aliveThreshold, err := strconv.Atoi(serviceTransport.AliveThreshold)
		if err != nil {
			return nil, err
		}

		deadThreshold, err := strconv.Atoi(serviceTransport.DeadThreshold)
		if err != nil {
			return nil, err
		}

		applicationServers, err := convertApplicationServersTransportToDomainModel(serviceTransport.ApplicationServersTransport,
			responseTimer,
			aliveThreshold,
			deadThreshold)
		if err != nil {
			return nil, err
		}

		serviceInfo := &domain.ServiceInfo{
			Address:            serviceTransport.IP + ":" + serviceTransport.Port,
			IP:                 serviceTransport.IP,
			Port:               serviceTransport.Port,
			IsUp:               false,
			BalanceType:        serviceTransport.BalanceType,
			RoutingType:        serviceTransport.RoutingType,
			Protocol:           serviceTransport.Protocol,
			Quorum:             quorum,
			HealthcheckType:    serviceTransport.HealthcheckType,
			HelloTimer:         helloTimer,
			ResponseTimer:      responseTimer,
			HCNearFieldsMode:   false,
			HCUserDefinedData:  nil,
			AliveThreshold:     aliveThreshold,
			DeadThreshold:      deadThreshold,
			ApplicationServers: applicationServers,
			HCStop:             make(chan struct{}, 1),
			HCStopped:          make(chan struct{}, 1),
		}
		servicesInfo = append(servicesInfo, serviceInfo)
	}

	return servicesInfo, nil
}

func convertApplicationServersTransportToDomainModel(applicationServersTransport []*ApplicationServerTransport,
	responseTimer time.Duration,
	aliveThreshold, deadThreshold int) (map[string]*domain.ApplicationServer, error) {
	applicationServers := make(map[string]*domain.ApplicationServer, len(applicationServersTransport))
	for _, applicationServerTransport := range applicationServersTransport {
		internalHC := domain.InternalHC{}
		ip, _, err := net.ParseCIDR(applicationServerTransport.IP + "/32")
		if err != nil {
			return nil, err
		}
		internalHC.HealthcheckType = applicationServerTransport.HealthcheckAddress
		internalHC.HealthcheckAddress = applicationServerTransport.HealthcheckAddress
		internalHC.ResponseTimer = responseTimer
		internalHC.LastIndexForAlive = 0
		internalHC.LastIndexForDead = 0
		internalHC.Mark = int(ip4toInt(ip))
		internalHC.NearFieldsMode = false
		internalHC.UserDefinedData = nil
		internalHC.AliveThreshold = make([]bool, aliveThreshold)
		internalHC.DeadThreshold = make([]bool, deadThreshold)

		applicationServer := &domain.ApplicationServer{
			Address:            applicationServerTransport.IP + ":" + applicationServerTransport.Port,
			IP:                 applicationServerTransport.IP,
			Port:               applicationServerTransport.Port,
			IsUp:               false,
			HealthcheckAddress: applicationServerTransport.HealthcheckAddress,
			InternalHC:         internalHC,
		}
		applicationServers[applicationServer.Address] = applicationServer
	}
	return applicationServers, nil
}

func ip4toInt(ipv4Address net.IP) int64 {
	ipv4Int := big.NewInt(0)
	ipv4Int.SetBytes(ipv4Address.To4())
	return ipv4Int.Int64()
}
