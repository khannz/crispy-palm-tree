package application

import (
	"context"
	"fmt"
	"log"
	"net"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// GrpcServer is used to implement portadapter.HCGetService.
type GrpcServer struct {
	address string
	facade  *IPVSFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedIPVSWokerServer
}

func NewGrpcServer(address string,
	facade *IPVSFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		address: address,
		facade:  facade,
		logging: logging,
	}
}

func convertPbApplicationServersToInternal(pbApplicationServers map[string]uint32) map[string]uint16 {
	internalApplicationServers := make(map[string]uint16, len(pbApplicationServers))
	for k, v := range pbApplicationServers {
		internalApplicationServers[k] = uint16(v)
	}
	return internalApplicationServers
}

// NewIPVSService implements portadapter.NewIPVSService
func (gs *GrpcServer) NewIPVSService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)
	return &transport.EmptyIpvsData{}, gs.facade.NewIPVSService(incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	)
}

// AddIPVSApplicationServersForService implements portadapter.AddIPVSApplicationServersForService
func (gs *GrpcServer) AddIPVSApplicationServersForService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)
	return &transport.EmptyIpvsData{}, gs.facade.AddIPVSApplicationServersForService(incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	)
}

// RemoveIPVSService implements portadapter.RemoveIPVSService
func (gs *GrpcServer) RemoveIPVSService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	return &transport.EmptyIpvsData{}, gs.facade.RemoveIPVSService(incomeIPVSService.Vip,
		convertedPort,
		convertedProtocol,
		incomeIPVSService.Id,
	)
}

// RemoveIPVSApplicationServersFromService implements portadapter.RemoveIPVSApplicationServersFromService
func (gs *GrpcServer) RemoveIPVSApplicationServersFromService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)
	return &transport.EmptyIpvsData{}, gs.facade.RemoveIPVSApplicationServersFromService(incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	)
}

// IsIPVSApplicationServerInService implements portadapter.IsIPVSApplicationServerInService
func (gs *GrpcServer) IsIPVSApplicationServerInService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.BoolData, error) {
	convertedPort := uint16(incomeIPVSService.Port)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)
	isIn, err := gs.facade.IsIPVSApplicationServerInService(
		incomeIPVSService.Vip,
		convertedPort,
		convertedApplicationServers,
		incomeIPVSService.Id,
	)
	if err != nil {
		return &transport.BoolData{IsIn: isIn}, err
	}
	return &transport.BoolData{IsIn: isIn}, nil
}

// IPVSFlush implements portadapter.IPVSFlush
func (gs *GrpcServer) IPVSFlush(ctx context.Context, incomeIPVSService *transport.EmptyIpvsData) (*transport.EmptyIpvsData, error) {
	return &transport.EmptyIpvsData{}, gs.facade.IPVSFlush()
}

func (grpcServer *GrpcServer) StartServer() error {
	lis, err := net.Listen("tcp", grpcServer.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterIPVSWokerServer(grpcServer.grpcSrv, grpcServer)
	go grpcServer.Serve(lis)
	return nil
}

func (grpcServer *GrpcServer) Serve(lis net.Listener) {
	if err := grpcServer.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (grpcServer *GrpcServer) CloseServer() {
	grpcServer.grpcSrv.Stop()
}
