package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/golang/protobuf/ptypes"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/grpc-healthcheck"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const grpcJobName = "healthcheck grpc job"

// GrpcServer is used to implement portadapter.HCGetService.
type GrpcServer struct {
	address string
	facade  *HCFacade
	grpcSrv *grpc.Server
	logging *logrus.Logger
	transport.UnimplementedHealthcheckWorkerServer
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

func (gs *GrpcServer) IsHttpAdvancedCheckOk(ctx context.Context, incomeHttpAdvancedCheck *transport.HttpAdvancedData) (*transport.IsOk, error) {
	id := incomeHttpAdvancedCheck.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job http adv check server %v", incomeHttpAdvancedCheck)

	timeout, _ := ptypes.Duration(incomeHttpAdvancedCheck.Timeout)
	isOk := gs.facade.IsHttpAdvancedCheckOk(
		incomeHttpAdvancedCheck.HcType,
		incomeHttpAdvancedCheck.HealthcheckAddress,
		incomeHttpAdvancedCheck.NearFieldsMode,
		incomeHttpAdvancedCheck.UserDefinedData,
		timeout,
		int(incomeHttpAdvancedCheck.Fwmark),
		id)

	outPbData := &transport.IsOk{
		IsOk: isOk,
		Id:   id,
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job http adv check server %v", incomeHttpAdvancedCheck)

	return outPbData, nil
}

func (gs *GrpcServer) IsHttpsCheckOk(ctx context.Context, incomeHttpsCheck *transport.HttpsData) (*transport.IsOk, error) {
	id := incomeHttpsCheck.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job https check server %v", incomeHttpsCheck)

	timeout, _ := ptypes.Duration(incomeHttpsCheck.Timeout)
	isOk := gs.facade.IsHttpsCheckOk(
		incomeHttpsCheck.HealthcheckAddress,
		timeout,
		int(incomeHttpsCheck.Fwmark),
		id)

	outPbData := &transport.IsOk{
		IsOk: isOk,
		Id:   id,
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job https check server %v", incomeHttpsCheck)

	return outPbData, nil
}

func (gs *GrpcServer) IsIcmpCheckOk(ctx context.Context, incomeIcmpCheck *transport.IcmpData) (*transport.IsOk, error) {
	id := incomeIcmpCheck.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job icmp check server %v", incomeIcmpCheck)

	timeout, _ := ptypes.Duration(incomeIcmpCheck.Timeout)
	isOk := gs.facade.IsIcmpCheckOk(
		incomeIcmpCheck.IpS,
		int(incomeIcmpCheck.Seq),
		timeout,
		int(incomeIcmpCheck.Fwmark),
		incomeIcmpCheck.Id)

	outPbData := &transport.IsOk{
		IsOk: isOk,
		Id:   id,
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job icmp check server %v", incomeIcmpCheck)

	return outPbData, nil
}

func (gs *GrpcServer) IsTcpCheckOk(ctx context.Context, incomeTcpCheck *transport.TcpData) (*transport.IsOk, error) {
	id := incomeTcpCheck.Id
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("got job tcp check server %v", incomeTcpCheck)

	timeout, _ := ptypes.Duration(incomeTcpCheck.Timeout)
	isOk := gs.facade.IsTcpCheckOk(
		incomeTcpCheck.HealthcheckAddress,
		timeout,
		int(incomeTcpCheck.Fwmark),
		incomeTcpCheck.Id)

	outPbData := &transport.IsOk{
		IsOk: isOk,
		Id:   id,
	}
	gs.facade.Logging.WithFields(logrus.Fields{
		"entity":   grpcJobName,
		"event id": id,
	}).Infof("completed job tcp check server %v", incomeTcpCheck)

	return outPbData, nil
}

func (grpcServer *GrpcServer) StartServer() error {
	if err := grpcServer.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup socket info: %v", err)
	}

	// lis, err := net.Listen("unix", grpcServer.address)
	lis, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer.grpcSrv = grpc.NewServer()
	transport.RegisterHealthcheckWorkerServer(grpcServer.grpcSrv, grpcServer)
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
