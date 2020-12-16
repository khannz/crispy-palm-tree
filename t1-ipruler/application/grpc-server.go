package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	transport "github.com/khannz/crispy-palm-tree/t1-ipruler/grpc-ipruler"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const grpcRouteName = "ip-ruler"

// GrpcServer ...
type GrpcServer struct {
	addr    string
	facade  *RouteFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedIPRulerGetWorkerServer
}

func NewGrpcServer(addr string,
	facade *RouteFacade,
	logging *logrus.Logger) *GrpcServer {
	return &GrpcServer{
		addr:    addr,
		facade:  facade,
		logging: logging,
	}
}

// AddToIPRuler ...
func (gs *GrpcServer) AddToIPRuler(ctx context.Context, incomeIpData *transport.IpData) (*transport.EmptyGetIPRulerData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeIpData.Id,
	}).Infof("got job add to ip rule service %v", incomeIpData)
	if err := gs.facade.AddIPRule(incomeIpData.TunDestIP, incomeIpData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcRouteName,
			"event id": incomeIpData.Id,
		}).Errorf("failed job add to ip rule service %v", incomeIpData)
		return &transport.EmptyGetIPRulerData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeIpData.Id,
	}).Infof("completed job add to ip rule service %v", incomeIpData)
	return &transport.EmptyGetIPRulerData{}, nil
}

// RemoveFromIPRuler ...
func (gs *GrpcServer) RemoveFromIPRuler(ctx context.Context, incomeIpData *transport.IpData) (*transport.EmptyGetIPRulerData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeIpData.Id,
	}).Infof("got job remove from ip rule service %v", incomeIpData)
	if err := gs.facade.RemoveIPRule(incomeIpData.TunDestIP, incomeIpData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcRouteName,
			"event id": incomeIpData.Id,
		}).Errorf("failed job remove from ip rule service %v", incomeIpData)
		return &transport.EmptyGetIPRulerData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeIpData.Id,
	}).Infof("completed job remove from ip rule service %v", incomeIpData)
	return &transport.EmptyGetIPRulerData{}, nil
}

// GetIPRulerRuntime ...
func (gs *GrpcServer) GetIPRulerRuntime(ctx context.Context, incomeEmptyData *transport.EmptyGetIPRulerData) (*transport.GetIPRulerRuntimeData, error) {
	currentConfig, err := gs.facade.GetIPRuleRuntimeConfig(incomeEmptyData.Id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": incomeEmptyData.Id,
		}).Errorf("failed to get runtime config: %v", err)
		return nil, err
	}

	convertedCurrentConfig := convertMapToPbMap(currentConfig)

	pbCurrentConfig := &transport.GetIPRulerRuntimeData{
		Fwmarks: convertedCurrentConfig,
		Id:      incomeEmptyData.Id,
	}
	return pbCurrentConfig, nil
}

func convertMapToPbMap(currentConfig map[int]struct{}) map[int64]*transport.EmptyGetIPRulerData {
	convertedCurrentConfig := make(map[int64]*transport.EmptyGetIPRulerData, len(currentConfig))
	for cc := range currentConfig {
		convertedCurrentConfig[int64(cc)] = &transport.EmptyGetIPRulerData{}
	}
	return convertedCurrentConfig
}

func (grpcServer *GrpcServer) StartServer() error {
	if err := grpcServer.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup socket info: %v", err)
	}

	// lis, err := net.Listen("unix", grpcServer.addr)
	lis, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterIPRulerGetWorkerServer(grpcServer.grpcSrv, grpcServer)
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
