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
	transport.UnimplementedDummyWokerServer
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
func (gs *GrpcServer) AddToDummy(ctx context.Context, incomeDummyData *transport.IpData) (*transport.EmptyDummyData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("got job add to dummy service %v", incomeDummyData)
	if err := gs.facade.AddToDummy(incomeDummyData.Ip, incomeDummyData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcDummyName,
			"event id": incomeDummyData.Id,
		}).Errorf("failed job add to dummy service %v", incomeDummyData)
		return &transport.EmptyDummyData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("completed job add to dummy service %v", incomeDummyData)
	return &transport.EmptyDummyData{}, nil
}

// RemoveFromDummy implements portadapter.RemoveFromDummy
func (gs *GrpcServer) RemoveFromDummy(ctx context.Context, incomeDummyData *transport.IpData) (*transport.EmptyDummyData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("got job remove from dummy service %v", incomeDummyData)
	if err := gs.facade.RemoveFromDummy(incomeDummyData.Ip, incomeDummyData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcDummyName,
			"event id": incomeDummyData.Id,
		}).Errorf("failed job remove from dummy service %v", incomeDummyData)
		return &transport.EmptyDummyData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcDummyName,
		"event id": incomeDummyData.Id,
	}).Infof("completed job remove from dummy service %v", incomeDummyData)
	return &transport.EmptyDummyData{}, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	lis, err := net.Listen("tcp", grpcServer.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterDummyWokerServer(grpcServer.grpcSrv, grpcServer)
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
