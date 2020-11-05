package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	grpcIpvsName = "ipvs grpc"
	protocol     = "unix"
)

// UdsGrpcServer is used to implement portadapter.HCGetService.
type UdsGrpcServer struct {
	sockAddr string
	facade   *TunnelFacade
	grpcSrv  *grpc.Server
	logging  *logrus.Logger
	transport.UnimplementedTunnelsCreateServer
	transport.UnimplementedTunnelsRemoveServer
}

func NewUdsGrpcServer(sockPath string,
	sockName string,
	facade *TunnelFacade,
	logging *logrus.Logger) *UdsGrpcServer {
	return &UdsGrpcServer{
		sockAddr: sockPath + sockName,
		facade:   facade,
		logging:  logging,
	}
}

// CreateTunnels implements portadapter.CreateTunnels
func (gs *UdsGrpcServer) CreateTunnels(ctx context.Context, incomeCreateTunnels *transport.PbTunnelsInfo) (*transport.PbTunnelsInfo, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeCreateTunnels.Id,
	}).Infof("got job create tunnels %v", incomeCreateTunnels)

	tunnelsInfo := convertIncomePbTunnelsInfoToDomainTunnelsInfo(incomeCreateTunnels)

	if err := gs.facade.CreateTunnels(tunnelsInfo, incomeCreateTunnels.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeCreateTunnels.Id,
		}).Errorf("failed job create tunnels %v", incomeCreateTunnels)
		return incomeCreateTunnels, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeCreateTunnels.Id,
	}).Infof("completed job create tunnels %v", tunnelsInfo)

	pbUpdatedTunnelsInfo := convertDomainTunnelsInfoToIncomePbTunnelsInfo(tunnelsInfo, incomeCreateTunnels.Id)
	return pbUpdatedTunnelsInfo, nil
}

// RemoveTunnels implements portadapter.RemoveTunnels
func (gs *UdsGrpcServer) RemoveTunnels(ctx context.Context, incomeRemoveTunnels *transport.PbTunnelsInfo) (*transport.PbTunnelsInfo, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeRemoveTunnels.Id,
	}).Infof("got job remove tunnels %v", incomeRemoveTunnels)

	tunnelsInfo := convertIncomePbTunnelsInfoToDomainTunnelsInfo(incomeRemoveTunnels)

	if err := gs.facade.RemoveTunnels(tunnelsInfo, incomeRemoveTunnels.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeRemoveTunnels.Id,
		}).Errorf("failed job remove tunnels %v", incomeRemoveTunnels)
		return incomeRemoveTunnels, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeRemoveTunnels.Id,
	}).Infof("completed job remove tunnels %v", tunnelsInfo)

	pbUpdatedTunnelsInfo := convertDomainTunnelsInfoToIncomePbTunnelsInfo(tunnelsInfo, incomeRemoveTunnels.Id)
	return pbUpdatedTunnelsInfo, nil
}

// RemoveAllTunnels implements portadapter.RemoveAllTunnels
func (gs *UdsGrpcServer) RemoveAllTunnels(ctx context.Context, incomeRemoveAllTunnels *transport.PbTunnelsInfo) (*transport.EmptyTunnelData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeRemoveAllTunnels.Id,
	}).Infof("got job remove tunnels %v", incomeRemoveAllTunnels)

	tunnelsInfo := convertIncomePbTunnelsInfoToDomainTunnelsInfo(incomeRemoveAllTunnels)

	if err := gs.facade.RemoveAllTunnels(tunnelsInfo, incomeRemoveAllTunnels.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeRemoveAllTunnels.Id,
		}).Errorf("failed job remove tunnels %v", incomeRemoveAllTunnels)
		return &transport.EmptyTunnelData{}, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeRemoveAllTunnels.Id,
	}).Infof("completed job remove tunnels %v", tunnelsInfo)

	return &transport.EmptyTunnelData{}, nil
}

func convertIncomePbTunnelsInfoToDomainTunnelsInfo(pbTunnelsInfo *transport.PbTunnelsInfo) []*domain.TunnelForApplicationServer {
	domainTunnelsInfo := make([]*domain.TunnelForApplicationServer, 0, len(pbTunnelsInfo.PbTunnelInfo))
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
	pbTunnelsInfo := make([]*transport.PbTunnelInfo, 0, len(tunnelsInfo))
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

func (grpcServer *UdsGrpcServer) StartServer() error {
	if err := grpcServer.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup socket info: %v", err)
	}

	lis, err := net.Listen(protocol, grpcServer.sockAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterTunnelsCreateServer(grpcServer.grpcSrv, grpcServer)
	transport.RegisterTunnelsRemoveServer(grpcServer.grpcSrv, grpcServer)
	go grpcServer.Serve(lis)
	return nil
}

func (grpcServer *UdsGrpcServer) cleanup() error {
	if _, err := os.Stat(grpcServer.sockAddr); err == nil {
		if err := os.RemoveAll(grpcServer.sockAddr); err != nil {
			return err
		}
	}
	return nil
}

func (grpcServer *UdsGrpcServer) Serve(lis net.Listener) {
	if err := grpcServer.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (grpcServer *UdsGrpcServer) CloseServer() {
	grpcServer.grpcSrv.Stop()
}
