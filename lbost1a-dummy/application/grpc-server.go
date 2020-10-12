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
func (gs *GrpcServer) AddToDummy(ctx context.Context, incomeIP *transport.IpData) (*transport.EmptyDummyData, error) {
	return &transport.EmptyDummyData{}, gs.facade.AddToDummy(incomeIP.Ip)
}

// RemoveFromDummy implements portadapter.RemoveFromDummy
func (gs *GrpcServer) RemoveFromDummy(ctx context.Context, incomeIP *transport.IpData) (*transport.EmptyDummyData, error) {
	return &transport.EmptyDummyData{}, gs.facade.RemoveFromDummy(incomeIP.Ip)
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
