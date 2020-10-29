package application

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/grpc-transport"
)

func pbServiceToDomainHCService(pbService *transport.PbService) *domain.HCService {
	domainApplicationServers := pbApplicationServersToDomainApplicationServers(pbService.PbApplicationServers)
	alivedAppServersForUp := int(pbService.AlivedAppServersForUp)
	hcRepeat, _ := ptypes.Duration(pbService.HcRepeat)
	hcTimeout, _ := ptypes.Duration(pbService.HcTimeout)
	hcRetriesForUP := int(pbService.HcRetriesForUP)
	hcRetriesForDown := int(pbService.HcRetriesForDown)
	return &domain.HCService{
		Address:               pbService.Address,
		IP:                    pbService.Ip,
		Port:                  pbService.Port,
		IsUp:                  pbService.IsUp,
		BalanceType:           pbService.BalanceType,
		RoutingType:           pbService.RoutingType,
		Protocol:              pbService.Protocol,
		AlivedAppServersForUp: alivedAppServersForUp,
		HCType:                pbService.HcType,
		HCRepeat:              hcRepeat,
		HCTimeout:             hcTimeout,
		HCNearFieldsMode:      pbService.HcNearFieldsMode,
		HCUserDefinedData:     pbService.HcUserDefinedData,
		HCRetriesForUP:        hcRetriesForUP,
		HCRetriesForDown:      hcRetriesForDown,
		HCApplicationServers:  domainApplicationServers,
		HCStop:                make(chan struct{}),
		HCStopped:             make(chan struct{}),
	}
}

func pbApplicationServersToDomainApplicationServers(pbApplicationServers map[string]*transport.PbService_PbApplicationServer) map[string]*domain.HCApplicationServer {
	domainApplicationServers := make(map[string]*domain.HCApplicationServer, len(pbApplicationServers))
	for i := range pbApplicationServers {
		domainApplicationServers[i] = &domain.HCApplicationServer{
			Address:   pbApplicationServers[i].Address,
			IP:        pbApplicationServers[i].Ip,
			Port:      pbApplicationServers[i].Port,
			IsUp:      pbApplicationServers[i].IsUp,
			HCAddress: pbApplicationServers[i].HcAddress,
		}
	}
	return domainApplicationServers
}

func domainHCServiceToPbService(domainHCService *domain.HCService) *transport.PbService {
	pbApplicationServers := domainApplicationServersToPbApplicationServers(domainHCService.HCApplicationServers)
	return &transport.PbService{
		Address:               domainHCService.Address,
		Ip:                    domainHCService.IP,
		Port:                  domainHCService.Port,
		IsUp:                  domainHCService.IsUp,
		BalanceType:           domainHCService.BalanceType,
		RoutingType:           domainHCService.RoutingType,
		Protocol:              domainHCService.Protocol,
		AlivedAppServersForUp: int32(domainHCService.AlivedAppServersForUp),
		HcType:                domainHCService.HCType,
		HcRepeat:              ptypes.DurationProto(domainHCService.HCRepeat),
		HcTimeout:             ptypes.DurationProto(domainHCService.HCRepeat),
		HcUserDefinedData:     domainHCService.HCUserDefinedData,
		HcNearFieldsMode:      domainHCService.HCNearFieldsMode,
		HcRetriesForUP:        int32(domainHCService.HCRetriesForUP),
		HcRetriesForDown:      int32(domainHCService.HCRetriesForDown),
		PbApplicationServers:  pbApplicationServers,
	}
}

func domainApplicationServersToPbApplicationServers(domainApplicationServers map[string]*domain.HCApplicationServer) map[string]*transport.PbService_PbApplicationServer {
	pbApplicationServer := make(map[string]*transport.PbService_PbApplicationServer, len(domainApplicationServers))
	for i := range domainApplicationServers {
		pbApplicationServer[i] = &transport.PbService_PbApplicationServer{
			Address:   domainApplicationServers[i].Address,
			Ip:        domainApplicationServers[i].IP,
			Port:      domainApplicationServers[i].Port,
			IsUp:      domainApplicationServers[i].IsUp,
			HcAddress: domainApplicationServers[i].HCAddress,
		}
	}
	return pbApplicationServer
}
