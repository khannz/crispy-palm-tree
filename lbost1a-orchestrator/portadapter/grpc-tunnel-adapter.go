package portadapter

import (
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/grpc-transport"
)

func convertIncomePbTunnelsInfoToDomainTunnelsInfo(pbTunnelsInfo *transport.PbTunnelsInfo) []*domain.TunnelForApplicationServer {
	domainTunnelsInfo := make([]*domain.TunnelForApplicationServer, len(pbTunnelsInfo.PbTunnelInfo))
	for i, pbTunnel := range pbTunnelsInfo.PbTunnelInfo {
		domainTunnelInfo := &domain.TunnelForApplicationServer{
			ApplicationServerIP:   pbTunnel.Ip,
			SysctlConfFile:        pbTunnel.SysctlConfFile,
			TunnelName:            pbTunnel.TunnelName,
			ServicesToTunnelCount: int(pbTunnel.ServicesToTunnelCount),
		}
		domainTunnelsInfo[i] = domainTunnelInfo
	}
	return domainTunnelsInfo
}

func convertDomainTunnelsInfoToIncomePbTunnelsInfo(tunnelsInfo []*domain.TunnelForApplicationServer, id string) *transport.PbTunnelsInfo {
	pbTunnelsInfo := make([]*transport.PbTunnelInfo, len(tunnelsInfo))
	for i, domainTunnel := range tunnelsInfo {
		pbTunnelInfo := &transport.PbTunnelInfo{
			Ip:                    domainTunnel.ApplicationServerIP,
			SysctlConfFile:        domainTunnel.SysctlConfFile,
			TunnelName:            domainTunnel.TunnelName,
			ServicesToTunnelCount: int32(domainTunnel.ServicesToTunnelCount),
		}
		pbTunnelsInfo[i] = pbTunnelInfo
	}

	pbModel := &transport.PbTunnelsInfo{
		Id:           id,
		PbTunnelInfo: pbTunnelsInfo,
	}
	return pbModel
}
