package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-orch"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const grpcOrchName = "orch"

// GrpcServer ...
type GrpcServer struct {
	addr    string
	facade  *T1OrchFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedSendRuntimeServer
}

func NewGrpcServer(addr string,
	facade *T1OrchFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		addr:    addr,
		facade:  facade,
		logging: logging,
	}
}

// DummyRuntime ...
func (gs *GrpcServer) DummyRuntime(ctx context.Context,
	incomeDummyRuntimeData *transport.DummyRuntimeData) (*transport.EmptyDummyData, error) {
	// TODO: implement
	return &transport.EmptyDummyData{}, nil
}

// SendRouteRuntime ...
func (gs *GrpcServer) SendRouteRuntime(ctx context.Context,
	incomeRouteRuntimeData *transport.SendRouteRuntimeData) (*transport.EmptySendRouteData, error) {
	// TODO: implement
	return &transport.EmptySendRouteData{}, nil
}

// SendTunnelRuntime ...
func (gs *GrpcServer) SendTunnelRuntime(ctx context.Context,
	incomeDummyRuntimeData *transport.SendTunnelRuntimeData) (*transport.EmptySendTunnelData, error) {
	// TODO: implement
	return &transport.EmptySendTunnelData{}, nil
}

// IpRuleRuntime ...
func (gs *GrpcServer) IpRuleRuntime(ctx context.Context,
	incomeDummyRuntimeData *transport.IpRuleRuntimeData) (*transport.EmptyIpRuleData, error) {
	// TODO: implement
	return &transport.EmptyIpRuleData{}, nil
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
	transport.RegisterSendRuntimeServer(grpcServer.grpcSrv, grpcServer)
	go grpcServer.serve(lis)
	return nil
}

func (grpcServer *GrpcServer) serve(lis net.Listener) {
	if err := grpcServer.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to grpc serve: %v", err)
	}
}

func (grpcServer *GrpcServer) CloseServer() {
	grpcServer.grpcSrv.Stop()
	if err := grpcServer.cleanup(); err != nil {
		grpcServer.logging.WithFields(logrus.Fields{
			"entity": grpcOrchName,
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
