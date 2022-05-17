package consul_provider

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers/consul"
	"github.com/pkg/errors"
)

var (
	//ErrNonePayload  no payload to be pass into =update
	ErrNonePayload = errors.New("payload is none")

	//ErrUnsupportedPayload payload is unsupported by package
	ErrUnsupportedPayload = errors.New("unsupported payload")
)

type (
	ServiceTransportData        consul.ServiceTransportData
	applicationServerTransports []consul.ApplicationServerTransport
)

func (from ServiceTransportData) ToServiceInfoConf() (domain.ServiceInfoConf, error) {

	var source []consul.ServiceTransport
	switch t := from.Payload.(type) {
	case consul.ServicesPayload:
		source = t.Services
	case consul.NonePayload:
		return nil, ErrNonePayload
	default:
		return nil, ErrUnsupportedPayload
	}

	servicesInfo := make(domain.ServiceInfoConf, len(source))
	for _, serviceTransport := range source {
		var (
			err                error
			quorum             int
			helloTimer         time.Duration
			responseTimer      time.Duration
			aliveThreshold     int
			deadThreshold      int
			applicationServers domain.ApplicationServers
		)

		if quorum, err = strconv.Atoi(serviceTransport.Quorum); err != nil {
			return nil, err
		}

		if helloTimer, err = time.ParseDuration(serviceTransport.HelloTimer); err != nil {
			return nil, err
		}

		if responseTimer, err = time.ParseDuration(serviceTransport.ResponseTimer); err != nil {
			return nil, err
		}

		if aliveThreshold, err = strconv.Atoi(serviceTransport.AliveThreshold); err != nil {
			return nil, err
		}

		if deadThreshold, err = strconv.Atoi(serviceTransport.DeadThreshold); err != nil {
			return nil, err
		}

		applicationServers, err = applicationServerTransports(serviceTransport.ApplicationServersTransport).
			ApplicationServers(
				responseTimer,
				aliveThreshold,
				deadThreshold,
			)
		if err != nil {
			return nil, err
		}

		serviceInfo := domain.ServiceInfo{
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

		if strings.EqualFold(serviceTransport.HealthcheckType, "http") ||
			strings.EqualFold(serviceTransport.HealthcheckType, "https") {

			if serviceTransport.Uri == "" {
				serviceInfo.Uri = "/"
			} else {
				serviceInfo.Uri = serviceTransport.Uri
			}
			if len(serviceTransport.ValidResponseCodes) == 0 {
				serviceInfo.ValidResponseCodes = []int64{http.StatusOK}
			} else {
				serviceInfo.ValidResponseCodes = serviceTransport.ValidResponseCodes
			}
		}
		servicesInfo[serviceInfo.Address] = &serviceInfo
	}

	return servicesInfo, nil
}

func (from applicationServerTransports) ApplicationServers(responseTimer time.Duration, aliveThreshold, deadThreshold int) (domain.ApplicationServers, error) {
	applicationServers := make(domain.ApplicationServers, len(from))
	for _, applicationServerTransport := range from {
		ip, _, err := net.ParseCIDR(applicationServerTransport.IP + "/32")
		if err != nil {
			return nil, err
		}
		var internalHC domain.InternalHC
		internalHC.HealthcheckType = applicationServerTransport.HealthcheckAddress
		internalHC.HealthcheckAddress = applicationServerTransport.HealthcheckAddress
		internalHC.ResponseTimer = responseTimer
		internalHC.LastIndexForAlive = 0
		internalHC.LastIndexForDead = 0
		internalHC.Mark = int(domain.IPType(ip).Int())
		internalHC.NearFieldsMode = false
		internalHC.UserDefinedData = nil
		internalHC.AliveThreshold = make([]bool, aliveThreshold)
		internalHC.DeadThreshold = make([]bool, deadThreshold)

		applicationServer := domain.ApplicationServer{
			Address:            applicationServerTransport.IP + ":" + applicationServerTransport.Port,
			IP:                 applicationServerTransport.IP,
			Port:               applicationServerTransport.Port,
			IsUp:               false,
			HealthcheckAddress: applicationServerTransport.HealthcheckAddress,
			InternalHC:         internalHC,
		}
		applicationServers[applicationServer.Address] = &applicationServer
	}
	return applicationServers, nil
}
