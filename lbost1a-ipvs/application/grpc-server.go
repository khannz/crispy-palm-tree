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

const grpcIpvsName = "ipvs grpc"

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
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("got job new ipvs service %v", incomeIPVSService)
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)

	if err := gs.facade.NewIPVSService(incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeIPVSService.Id,
		}).Errorf("failed job new ipvs service %v", incomeIPVSService)
		return &transport.EmptyIpvsData{}, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job new ipvs service %v", incomeIPVSService)
	return &transport.EmptyIpvsData{}, nil
}

// AddIPVSApplicationServersForService implements portadapter.AddIPVSApplicationServersForService
func (gs *GrpcServer) AddIPVSApplicationServersForService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("got job add application servers to ipvs service %v", incomeIPVSService)
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)

	if err := gs.facade.AddIPVSApplicationServersForService(incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeIPVSService.Id,
		}).Errorf("failed job add application servers to ipvs service %v", incomeIPVSService)
		return &transport.EmptyIpvsData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job add application servers to ipvs service %v", incomeIPVSService)
	return &transport.EmptyIpvsData{}, nil
}

// RemoveIPVSService implements portadapter.RemoveIPVSService
func (gs *GrpcServer) RemoveIPVSService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("got job remove ipvs service %v", incomeIPVSService)
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)

	if err := gs.facade.RemoveIPVSService(incomeIPVSService.Vip,
		convertedPort,
		convertedProtocol,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeIPVSService.Id,
		}).Errorf("failed job remove ipvs service %v", incomeIPVSService)
		return &transport.EmptyIpvsData{}, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job remove ipvs service %v", incomeIPVSService)

	return &transport.EmptyIpvsData{}, nil
}

// RemoveIPVSApplicationServersFromService implements portadapter.RemoveIPVSApplicationServersFromService
func (gs *GrpcServer) RemoveIPVSApplicationServersFromService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.EmptyIpvsData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("got job remove application servers from ipvs service %v", incomeIPVSService)
	convertedPort := uint16(incomeIPVSService.Port)
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)

	if err := gs.facade.RemoveIPVSApplicationServersFromService(incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeIPVSService.Id,
		}).Errorf("failed job remove application servers from ipvs service %v", incomeIPVSService)
		return &transport.EmptyIpvsData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job remove application servers from ipvs service %v", incomeIPVSService)
	return &transport.EmptyIpvsData{}, nil
}

// IsIPVSApplicationServerInService implements portadapter.IsIPVSApplicationServerInService
func (gs *GrpcServer) IsIPVSApplicationServerInService(ctx context.Context, incomeIPVSService *transport.PbIPVSServices) (*transport.BoolData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("got job is application server in ipvs service %v", incomeIPVSService)
	convertedPort := uint16(incomeIPVSService.Port)
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)
	isIn, err := gs.facade.IsIPVSApplicationServerInService(
		incomeIPVSService.Vip,
		convertedPort,
		convertedApplicationServers,
		incomeIPVSService.Id,
	)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeIPVSService.Id,
		}).Errorf("failed job is application server in ipvs service %v", incomeIPVSService)
		return &transport.BoolData{IsIn: isIn}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job is application server in ipvs service %v", incomeIPVSService)
	return &transport.BoolData{IsIn: isIn}, nil
}

// IPVSFlush implements portadapter.IPVSFlush
func (gs *GrpcServer) IPVSFlush(ctx context.Context, incomeIPVSService *transport.EmptyIpvsData) (*transport.EmptyIpvsData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Info("got job flush ipvs")

	if err := gs.facade.IPVSFlush(); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": incomeIPVSService.Id,
		}).Error("failed job flush ipvs")
		return &transport.EmptyIpvsData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Info("completed job flush ipvs")
	return &transport.EmptyIpvsData{}, nil
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
