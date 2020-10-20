package application

import (
	"context"
	"fmt"
	"log"
	"net"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	id := incomePbService.Id
	outHCService, err := gs.facade.HCGetService(incomeHCService, id)
	if err != nil {
		return nil, err
	}
	outPbService := domainHCServiceToPbService(outHCService)
	outPbService.Id = id
	return outPbService, nil
}

// HCGetPbServiceS implements portadapter.HCGetPbServiceS
func (gs *GrpcServer) HCGetPbServiceS(ctx context.Context, empty *transport.EmptyHcData) (*transport.PbServices, error) {
	id := empty.Id
	outHCServices, err := gs.facade.HCGetServices(id)
	if err != nil {
		return nil, err
	}

	m := map[string]*transport.PbService{}
	outPbServices := &transport.PbServices{Services: m}
	for i := range outHCServices {
		outPbService := domainHCServiceToPbService(outHCServices[i])
		outPbServices.Services[outPbService.Address] = outPbService
	}
	outPbServices.Id = id
	return outPbServices, nil
}

// HCNewService implements portadapter.HCNewService
func (gs *GrpcServer) HCNewPbService(ctx context.Context, incomePbService *transport.PbService) (*transport.EmptyHcData, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	id := incomePbService.Id
	if err := gs.facade.HCNewService(incomeHCService, id); err != nil {
		return nil, err
	}
	return &transport.EmptyHcData{Id: id}, nil
}

func (gs *GrpcServer) HCUpdatePbService(ctx context.Context, incomePbService *transport.PbService) (*transport.PbService, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	id := incomePbService.Id
	outHCService, err := gs.facade.HCUpdateService(incomeHCService, id)
	if err != nil {
		return nil, err
	}
	outPbService := domainHCServiceToPbService(outHCService)
	outPbService.Id = id
	return outPbService, nil
}

func (gs *GrpcServer) HCRemovePbService(ctx context.Context, incomePbService *transport.PbService) (*transport.EmptyHcData, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	id := incomePbService.Id
	if err := gs.facade.HCRemoveService(incomeHCService, id); err != nil {
		return nil, err
	}
	return &transport.EmptyHcData{Id: id}, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	lis, err := net.Listen("tcp", grpcServer.address)
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

func (grpcServer *GrpcServer) Serve(lis net.Listener) {
	if err := grpcServer.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (grpcServer *GrpcServer) CloseServer() {
	grpcServer.grpcSrv.Stop()
}
