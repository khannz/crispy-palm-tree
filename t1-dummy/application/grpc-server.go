package application

import (
	"context"
	"fmt"
	"log"
	"net"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-dummy/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const grpcDummyName = "dummy"

// GrpcServer is used to implement portadapter.HCGetService.
type GrpcServer struct {
	port    string
	facade  *DummyFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedDummyGetWorkerServer
}

func NewGrpcServer(port string,
	facade *DummyFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		port:    port,
		facade:  facade,
		logging: logging,
	}
}

// AddToDummy implements portadapter.AddToDummy
func (gs *GrpcServer) AddToDummy(ctx context.Context, incomeDummyData *transport.IpData) (*transport.EmptyGetDummyData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("got job add to dummy service %v", incomeDummyData)
	if err := gs.facade.AddToDummy(incomeDummyData.Ip, incomeDummyData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcDummyName,
			"event id": incomeDummyData.Id,
		}).Errorf("failed job add to dummy service %v", incomeDummyData)
		return &transport.EmptyGetDummyData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("completed job add to dummy service %v", incomeDummyData)
	return &transport.EmptyGetDummyData{}, nil
}

// RemoveFromDummy implements portadapter.RemoveFromDummy
func (gs *GrpcServer) RemoveFromDummy(ctx context.Context, incomeDummyData *transport.IpData) (*transport.EmptyGetDummyData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("got job remove from dummy service %v", incomeDummyData)
	if err := gs.facade.RemoveFromDummy(incomeDummyData.Ip, incomeDummyData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcDummyName,
			"event id": incomeDummyData.Id,
		}).Errorf("failed job remove from dummy service %v", incomeDummyData)
		return &transport.EmptyGetDummyData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("completed job remove from dummy service %v", incomeDummyData)
	return &transport.EmptyGetDummyData{}, nil
}

// DummyGetRuntime ...
func (gs *GrpcServer) DummyGetRuntime(ctx context.Context, incomeEmptyData *transport.EmptyGetDummyData) (*transport.GetDummyRuntimeData, error) {
	currentConfig, err := gs.facade.GetDummyRuntimeConfig(incomeEmptyData.Id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": incomeEmptyData.Id,
		}).Errorf("failed to get runtime config: %v", err)
		return nil, err
	}
	pbCurrentConfig := convertRuntimeConfigToPbRuntimeConfig(currentConfig, incomeEmptyData.Id)
	return pbCurrentConfig, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	lis, err := net.Listen("tcp", grpcServer.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterDummyGetWorkerServer(grpcServer.grpcSrv, grpcServer)
	go grpcServer.Serve(lis)
	return nil
}

func (grpcServer *GrpcServer) Serve(lis net.Listener) {
	if err := grpcServer.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to grpc serve: %v", err)
	}
}

func (grpcServer *GrpcServer) CloseServer() {
	grpcServer.grpcSrv.Stop()
}

func convertRuntimeConfigToPbRuntimeConfig(runtimeConfig map[string]struct{}, id string) *transport.GetDummyRuntimeData {
	ed := &transport.EmptyGetDummyData{}
	pbMap := make(map[string]*transport.EmptyGetDummyData)
	for k := range runtimeConfig {
		pbMap[k] = ed
	}

	return &transport.GetDummyRuntimeData{
		Services: pbMap,
		Id:       id,
	}
}
