package portadapter

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/khannz/crispy-palm-tree/lbost1a-controller/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-controller/grpc-transport"
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

func pbApplicationServersToDomainApplicationServers(pbApplicationServers []*transport.PbService_PbApplicationServer) []*domain.ApplicationServer {
	domainApplicationServers := make([]*domain.ApplicationServer, len(pbApplicationServers))
	for i := range pbApplicationServers {
		domainApplicationServers[i] = &domain.ApplicationServer{
			Address:   pbApplicationServers[i].Address,
			IP:        pbApplicationServers[i].Ip,
			Port:      pbApplicationServers[i].Port,
			IsUp:      pbApplicationServers[i].IsUp,
			HCAddress: pbApplicationServers[i].HcAddress,
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

func domainApplicationServersToPbApplicationServers(domainApplicationServers []*domain.ApplicationServer) []*transport.PbService_PbApplicationServer {
	pbApplicationServer := make([]*transport.PbService_PbApplicationServer, len(domainApplicationServers))
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
