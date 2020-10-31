package portadapter

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/grpc-transport"
)

func pbServiceToDomainServiceInfo(pbService *transport.PbService) *domain.ServiceInfo {
	domainApplicationServers := pbApplicationServersToDomainApplicationServers(pbService.PbApplicationServers)
	alivedAppServersForUp := int(pbService.AlivedAppServersForUp)
	hcRepeat, _ := ptypes.Duration(pbService.HcRepeat)
	hcTimeout, _ := ptypes.Duration(pbService.HcTimeout)
	hcRetriesForUP := int(pbService.HcRetriesForUP)
	hcRetriesForDown := int(pbService.HcRetriesForDown)
	return &domain.ServiceInfo{
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
		ApplicationServers:    domainApplicationServers,
		HCStop:                make(chan struct{}),
		HCStopped:             make(chan struct{}),
	}
}

func pbApplicationServersToDomainApplicationServers(pbApplicationServers map[string]*transport.PbService_PbApplicationServer) map[string]*domain.ApplicationServer {
	domainApplicationServers := make(map[string]*domain.ApplicationServer, len(pbApplicationServers))
	for k := range pbApplicationServers {
		domainApplicationServers[k] = &domain.ApplicationServer{
			Address:   pbApplicationServers[k].Address,
			IP:        pbApplicationServers[k].Ip,
			Port:      pbApplicationServers[k].Port,
			IsUp:      pbApplicationServers[k].IsUp,
			HCAddress: pbApplicationServers[k].HcAddress,
		}
	}
	return domainApplicationServers
}

func domainServiceInfoToPbService(domainServiceInfo *domain.ServiceInfo) *transport.PbService {
	pbApplicationServers := domainApplicationServersToPbApplicationServers(domainServiceInfo.ApplicationServers)
	return &transport.PbService{
		Address:               domainServiceInfo.Address,
		Ip:                    domainServiceInfo.IP,
		Port:                  domainServiceInfo.Port,
		IsUp:                  domainServiceInfo.IsUp,
		BalanceType:           domainServiceInfo.BalanceType,
		RoutingType:           domainServiceInfo.RoutingType,
		Protocol:              domainServiceInfo.Protocol,
		AlivedAppServersForUp: int32(domainServiceInfo.AlivedAppServersForUp),
		HcType:                domainServiceInfo.HCType,
		HcRepeat:              ptypes.DurationProto(domainServiceInfo.HCRepeat),
		HcTimeout:             ptypes.DurationProto(domainServiceInfo.HCRepeat),
		HcUserDefinedData:     domainServiceInfo.HCUserDefinedData,
		HcNearFieldsMode:      domainServiceInfo.HCNearFieldsMode,
		HcRetriesForUP:        int32(domainServiceInfo.HCRetriesForUP),
		HcRetriesForDown:      int32(domainServiceInfo.HCRetriesForDown),
		PbApplicationServers:  pbApplicationServers,
	}
}

func domainApplicationServersToPbApplicationServers(domainApplicationServers map[string]*domain.ApplicationServer) map[string]*transport.PbService_PbApplicationServer {
	pbApplicationServer := make(map[string]*transport.PbService_PbApplicationServer, len(domainApplicationServers))
	for k := range domainApplicationServers {
		pbApplicationServer[k] = &transport.PbService_PbApplicationServer{
			Address:   domainApplicationServers[k].Address,
			Ip:        domainApplicationServers[k].IP,
			Port:      domainApplicationServers[k].Port,
			IsUp:      domainApplicationServers[k].IsUp,
			HcAddress: domainApplicationServers[k].HCAddress,
		}
	}
	return pbApplicationServer
}
