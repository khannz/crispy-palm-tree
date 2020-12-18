package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/grpc-tunnel"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const grpcRouteName = "tunnel"

// GrpcServer ...
type GrpcServer struct {
	addr    string
	facade  *TunnelFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedTunnelGetWorkerServer
}

func NewGrpcServer(addr string,
	facade *TunnelFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		addr:    addr,
		facade:  facade,
		logging: logging,
	}
}

// AddToRoute ...
func (gs *GrpcServer) AddTunnel(ctx context.Context, incomeTunnelData *transport.TunnelData) (*transport.EmptyTunnelData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeTunnelData.Id,
	}).Infof("got job add to tunnel service %v", incomeTunnelData)
	if err := gs.facade.AddTunnel(incomeTunnelData.HcTunDestIP, incomeTunnelData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcRouteName,
			"event id": incomeTunnelData.Id,
		}).Errorf("failed job add to tunnel service %v", incomeTunnelData)
		return &transport.EmptyTunnelData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeTunnelData.Id,
	}).Infof("completed job add to tunnel service %v", incomeTunnelData)
	return &transport.EmptyTunnelData{}, nil
}

// RemoveTunnel ...
func (gs *GrpcServer) RemoveTunnel(ctx context.Context, incomeTunnelData *transport.TunnelData) (*transport.EmptyTunnelData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeTunnelData.Id,
	}).Infof("got job remove from tunnel service %v", incomeTunnelData)
	if err := gs.facade.RemoveTunnel(incomeTunnelData.HcTunDestIP, incomeTunnelData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcRouteName,
			"event id": incomeTunnelData.Id,
		}).Errorf("failed job remove from tunnel service %v", incomeTunnelData)
		return &transport.EmptyTunnelData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeTunnelData.Id,
	}).Infof("completed job remove from tunnel service %v", incomeTunnelData)
	return &transport.EmptyTunnelData{}, nil
}

// GetTunnelRuntime ...
func (gs *GrpcServer) GetTunnelRuntime(ctx context.Context, incomeEmptyData *transport.EmptyTunnelData) (*transport.GetTunnelRuntimeData, error) {
	currentConfig, err := gs.facade.GetTunnelRuntime(incomeEmptyData.Id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": incomeEmptyData.Id,
		}).Errorf("failed to get runtime config: %v", err)
		return nil, err
	}

	convertedRuntimeConfig := convertRuntimeConfig(currentConfig)
	pbCurrentConfig := &transport.GetTunnelRuntimeData{
		Tunnels: convertedRuntimeConfig,
		Id:      incomeEmptyData.Id,
	}
	return pbCurrentConfig, nil
}

func convertRuntimeConfig(runtimeConfig map[string]struct{}) map[string]int32 {
	convertedRuntimeConfig := make(map[string]int32, len(runtimeConfig))
	for k := range runtimeConfig {
		convertedRuntimeConfig[k] = 0
	}
	return convertedRuntimeConfig
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
	transport.RegisterTunnelGetWorkerServer(grpcServer.grpcSrv, grpcServer)
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
