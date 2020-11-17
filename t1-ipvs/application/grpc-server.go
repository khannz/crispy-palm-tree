package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const grpcIpvsName = "ipvs grpc"

// GrpcServer is used to implement portadapter.HCGetService.
type GrpcServer struct {
	addr    string
	facade  *IPVSFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedIPVSGetWokerServer
}

func NewGrpcServer(addr string,
	facade *IPVSFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		addr:    addr,
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

func convertInternalToPbApplicationServers(internalApplicationServers map[string]uint16) map[string]uint32 {
	pbApplicationServers := make(map[string]uint32, len(internalApplicationServers))
	for k, v := range internalApplicationServers {
		pbApplicationServers[k] = uint32(v)
	}
	return pbApplicationServers
}

func convertInternalServicesToPbServices(internalServices map[string]map[string]uint16, id string) *transport.PbGetIPVSRawServicesData {
	pbServices := make(map[string]*transport.PbGetRawIPVSServiceData, len(internalServices))
	for k, v := range internalServices {
		applicationServers := &transport.PbGetRawIPVSServiceData{
			RawServiceData: convertInternalToPbApplicationServers(v),
		}
		pbServices[k] = applicationServers
	}

	return &transport.PbGetIPVSRawServicesData{
		RawServicesData: pbServices,
		Id:              id,
	}
}

// NewIPVSService implements portadapter.NewIPVSService
func (gs *GrpcServer) NewIPVSService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
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
		return &transport.EmptyGetIPVSData{}, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job new ipvs service %v", incomeIPVSService)
	return &transport.EmptyGetIPVSData{}, nil
}

// AddIPVSApplicationServersForService implements portadapter.AddIPVSApplicationServersForService
func (gs *GrpcServer) AddIPVSApplicationServersForService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
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
		return &transport.EmptyGetIPVSData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job add application servers to ipvs service %v", incomeIPVSService)
	return &transport.EmptyGetIPVSData{}, nil
}

// RemoveIPVSService implements portadapter.RemoveIPVSService
func (gs *GrpcServer) RemoveIPVSService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
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
		return &transport.EmptyGetIPVSData{}, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job remove ipvs service %v", incomeIPVSService)

	return &transport.EmptyGetIPVSData{}, nil
}

// RemoveIPVSApplicationServersFromService implements portadapter.RemoveIPVSApplicationServersFromService
func (gs *GrpcServer) RemoveIPVSApplicationServersFromService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
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
		return &transport.EmptyGetIPVSData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": incomeIPVSService.Id,
	}).Infof("completed job remove application servers from ipvs service %v", incomeIPVSService)
	return &transport.EmptyGetIPVSData{}, nil
}

// GetIPVSRuntime implements portadapter.GetIPVSRuntime
func (gs *GrpcServer) GetIPVSRuntime(ctx context.Context, emptyIPVSService *transport.EmptyGetIPVSData) (*transport.PbGetIPVSRawServicesData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": emptyIPVSService.Id,
	}).Info("got job is get IPVS runtime config")
	ipvsRuntime, err := gs.facade.GetIPVSRuntime(emptyIPVSService.Id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcIpvsName,
			"event id": emptyIPVSService.Id,
		}).Errorf("failed job is get IPVS runtime config: %v", err)
		return nil, err
	}

	pbIPVSServices := convertInternalServicesToPbServices(ipvsRuntime, emptyIPVSService.Id)
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcIpvsName,
		"event id": emptyIPVSService.Id,
	}).Infof("completed job is get IPVS runtime config")

	return pbIPVSServices, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	if err := grpcServer.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup socket info: %v", err)
	}

	lis, err := net.Listen("unix", grpcServer.addr)
	// lis, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterIPVSGetWokerServer(grpcServer.grpcSrv, grpcServer)
	go grpcServer.serve(lis)
	return nil
}

func (grpcServer *GrpcServer) serve(lis net.Listener) {
	if err := grpcServer.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (grpcServer *GrpcServer) CloseServer() {
	grpcServer.grpcSrv.Stop()
	if err := grpcServer.cleanup(); err != nil {
		grpcServer.logging.WithFields(logrus.Fields{
			"entity": sendRuntimeConfigName,
		}).Errorf("failed to cleanup grpc: %v", err)
	}
}

func (grpcServer *GrpcServer) cleanup() error {
	if _, err := os.Stat(grpcServer.addr); err == nil {
		if err := os.RemoveAll(grpcServer.addr); err != nil {
			return err
		}
	}
	return nil
}
