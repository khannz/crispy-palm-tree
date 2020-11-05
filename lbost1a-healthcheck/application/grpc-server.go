package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	grpcJobName = "healthcheck grpc job"
	protocol    = "unix"
)

// GrpcServer is used to implement portadapter.HCGetService.
type GrpcServer struct {
	address string
	facade  *HCFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedHCGetServer
	transport.UnimplementedHCNewServer
	transport.UnimplementedHCUpdateServer
}

func NewGrpcServer(address string,
	facade *HCFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		address: address,
		facade:  facade,
		logging: logging,
	}
}

// HCGetPbService implements portadapter.HCGetPbService
func (gs *GrpcServer) HCGetPbService(ctx context.Context, incomePbService *transport.PbService) (*transport.PbService, error) {
	id := incomePbService.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job get hc service %v", incomePbService)
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	outHCService, err := gs.facade.HCGetService(incomeHCService, id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcJobName,
			"event id": id,
		}).Errorf("failed job get hc service %v", incomePbService)
		return nil, err
	}
	outPbService := domainHCServiceToPbService(outHCService)
	outPbService.Id = id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job get hc service %v", incomePbService)
	return outPbService, nil
}

// HCGetPbServiceS implements portadapter.HCGetPbServiceS
func (gs *GrpcServer) HCGetPbServiceS(ctx context.Context, empty *transport.EmptyHcData) (*transport.PbServices, error) {
	id := empty.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Info("got job get hc services")
	outHCServices, err := gs.facade.HCGetServices(id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcJobName,
			"event id": id,
		}).Errorf("failed job get hc services")
		return nil, err
	}

	m := map[string]*transport.PbService{}
	outPbServices := &transport.PbServices{Services: m}
	for i := range outHCServices {
		outPbService := domainHCServiceToPbService(outHCServices[i])
		outPbServices.Services[outPbService.Address] = outPbService
	}
	outPbServices.Id = id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job get hc services")
	return outPbServices, nil
}

// HCNewService implements portadapter.HCNewService
func (gs *GrpcServer) HCNewPbService(ctx context.Context, incomePbService *transport.PbService) (*transport.EmptyHcData, error) {
	id := incomePbService.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job new hc service %v", incomePbService)
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	if err := gs.facade.HCNewService(incomeHCService, id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcJobName,
			"event id": id,
		}).Errorf("failed job new hc service %v", incomePbService)
		return nil, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job new hc service %v", incomePbService)
	return &transport.EmptyHcData{Id: id}, nil
}

func (gs *GrpcServer) HCUpdatePbService(ctx context.Context, incomePbService *transport.PbService) (*transport.PbService, error) {
	id := incomePbService.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job update hc service %v", incomePbService)
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	outHCService, err := gs.facade.HCUpdateService(incomeHCService, id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcJobName,
			"event id": id,
		}).Errorf("failed job update hc service %v", incomePbService)
		return nil, err
	}
	outPbService := domainHCServiceToPbService(outHCService)
	outPbService.Id = id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job update hc service %v", incomePbService)
	return outPbService, nil
}

func (gs *GrpcServer) HCRemovePbService(ctx context.Context, incomePbService *transport.PbService) (*transport.EmptyHcData, error) {
	id := incomePbService.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job remove hc service %v", incomePbService)
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	if err := gs.facade.HCRemoveService(incomeHCService, id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcJobName,
			"event id": id,
		}).Errorf("failed job remove hc service %v", incomePbService)
		return nil, err
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job remove hc service %v", incomePbService)
	return &transport.EmptyHcData{Id: id}, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	if err := grpcServer.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup socket info: %v", err)
	}

	lis, err := net.Listen(protocol, grpcServer.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterHCGetServer(grpcServer.grpcSrv, grpcServer)
	transport.RegisterHCNewServer(grpcServer.grpcSrv, grpcServer)
	transport.RegisterHCUpdateServer(grpcServer.grpcSrv, grpcServer)
	go grpcServer.Serve(lis)
	return nil
}

func (grpcServer *GrpcServer) cleanup() error {
	if _, err := os.Stat(grpcServer.address); err == nil {
		if err := os.RemoveAll(grpcServer.address); err != nil {
			return err
		}
	}
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
