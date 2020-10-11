package application

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// GrpcServer is used to implement portadapter.HCGetService.
type GrpcServer struct {
	address string
	facade  *HCFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	UnimplementedHCGetServer
	UnimplementedHCNewServer
	UnimplementedHCUpdateServer
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
func (gs *GrpcServer) HCGetPbService(ctx context.Context, incomePbService *PbService) (*PbService, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	outHCService, err := gs.facade.HCGetService(incomeHCService)
	if err != nil {
		return nil, err
	}
	outPbService := domainHCServiceToPbService(outHCService)
	return outPbService, nil
}

// HCGetPbServiceS implements portadapter.HCGetPbServiceS
func (gs *GrpcServer) HCGetPbServiceS(ctx context.Context, empty *EmptyPbService) (*PbServices, error) {
	outHCServices, err := gs.facade.HCGetServices()
	if err != nil {
		return nil, err
	}

	m := map[string]*PbService{}
	outPbServices := &PbServices{Services: m}
	for i := range outHCServices {
		outPbService := domainHCServiceToPbService(outHCServices[i])
		outPbServices.Services[outPbService.Address] = outPbService
	}
	return outPbServices, nil
}

// HCNewService implements portadapter.HCNewService
func (gs *GrpcServer) HCNewPbService(ctx context.Context, incomePbService *PbService) (*EmptyPbService, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	if err := gs.facade.HCNewService(incomeHCService); err != nil {
		return nil, err
	}
	return &EmptyPbService{}, nil
}

func (gs *GrpcServer) HCUpdatePbService(ctx context.Context, incomePbService *PbService) (*PbService, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	outHCService, err := gs.facade.HCUpdateService(incomeHCService)
	if err != nil {
		return nil, err
	}
	outPbService := domainHCServiceToPbService(outHCService)
	return outPbService, nil
}

func (gs *GrpcServer) HCRemovePbService(ctx context.Context, incomePbService *PbService) (*EmptyPbService, error) {
	incomeHCService := pbServiceToDomainHCService(incomePbService)
	if err := gs.facade.HCRemoveService(incomeHCService); err != nil {
		return nil, err
	}
	return &EmptyPbService{}, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	lis, err := net.Listen("tcp", grpcServer.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	RegisterHCGetServer(grpcServer.grpcSrv, grpcServer)
	RegisterHCNewServer(grpcServer.grpcSrv, grpcServer)
	RegisterHCUpdateServer(grpcServer.grpcSrv, grpcServer)
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
