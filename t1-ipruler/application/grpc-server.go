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
	transport.UnimplementedRouteGetWorkerServer
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

// AddToRoute ...
func (gs *GrpcServer) AddIPRule(ctx context.Context, incomeRouteData *transport.RouteData) (*transport.EmptyRouteData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeRouteData.Id,
	}).Infof("got job add to ip rule service %v", incomeRouteData)
	if err := gs.facade.AddIPRule(incomeRouteData.HcDestIP, incomeRouteData.HcTunDestIP, incomeRouteData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcRouteName,
			"event id": incomeRouteData.Id,
		}).Errorf("failed job add to ip rule service %v", incomeRouteData)
		return &transport.EmptyRouteData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeRouteData.Id,
	}).Infof("completed job add to ip rule service %v", incomeRouteData)
	return &transport.EmptyRouteData{}, nil
}

// RemoveIPRule ...
func (gs *GrpcServer) RemoveIPRule(ctx context.Context, incomeRouteData *transport.RouteData) (*transport.EmptyRouteData, error) {
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeRouteData.Id,
	}).Infof("got job remove from ip rule service %v", incomeRouteData)
	if err := gs.facade.RemoveIPRule(incomeRouteData.HcDestIP, incomeRouteData.HcTunDestIP, incomeRouteData.NeedRemoveTunnel, incomeRouteData.Id); err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   grpcRouteName,
			"event id": incomeRouteData.Id,
		}).Errorf("failed job remove from ip rule service %v", incomeRouteData)
		return &transport.EmptyRouteData{}, err
	}

	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcRouteName,
		"event id": incomeRouteData.Id,
	}).Infof("completed job remove from ip rule service %v", incomeRouteData)
	return &transport.EmptyRouteData{}, nil
}

// GetIPRuleRuntimeConfig ...
func (gs *GrpcServer) GetIPRuleRuntimeConfig(ctx context.Context, incomeEmptyData *transport.EmptyRouteData) (*transport.GetAllRoutesData, error) {
	currentConfig, err := gs.facade.GetIPRuleRuntimeConfig(incomeEmptyData.Id)
	if err != nil {
		gs.facade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": incomeEmptyData.Id,
		}).Errorf("failed to get runtime config: %v", err)
		return nil, err
	}
	pbCurrentConfig := &transport.GetAllRoutesData{
		RouteData: currentConfig,
		Id:        incomeEmptyData.Id,
	}
	return pbCurrentConfig, nil
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
	transport.RegisterRouteGetWorkerServer(grpcServer.grpcSrv, grpcServer)
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
