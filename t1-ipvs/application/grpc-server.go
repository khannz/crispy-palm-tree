package application

import (
	"context"
	"net"
	"os"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/grpc-ipvs"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	addr    string
	facade  *IPVSFacade
	grpcSrv *grpc.Server
	logger  *zerolog.Logger
	transport.UnimplementedIPVSGetWorkerServer
}

func NewGrpcServer(
	addr string,
	facade *IPVSFacade,
	logger *zerolog.Logger,
) *GrpcServer {
	return &GrpcServer{
		addr:   addr,
		facade: facade,
		logger: logger,
	}
}

func (gs *GrpcServer) NewIPVSService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
	//gs.facade.Logger.Info().
	//	Uint32("protocol", incomeIPVSService.Protocol).
	//	Str("vip", incomeIPVSService.Vip).
	//	Uint32("port", incomeIPVSService.Port).
	//	Str("balanceType", incomeIPVSService.BalanceType).
	//	Msg("creating vip")
	// FIXME why port comes in uint32?
	convertedPort := uint16(incomeIPVSService.Port)
	// FIXME why proto comes in uint32?
	convertedProtocol := uint16(incomeIPVSService.Protocol)

	if err := gs.facade.NewIPVSService(
		incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logger.Error().Err(err).Msg("failed to create vip")
		return &transport.EmptyGetIPVSData{}, err
	}
	return &transport.EmptyGetIPVSData{}, nil
}

func (gs *GrpcServer) AddIPVSApplicationServersForService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
	//for realAddr, realPort := range incomeIPVSService.ApplicationServers {
	//	gs.facade.Logger.Info().
	//		Str("addr", realAddr).
	//		Uint32("port", realPort).
	//		Msg("adding real to vip")
	//}

	// FIXME why it comes in uint32?
	convertedPort := uint16(incomeIPVSService.Port)
	// FIXME why it comes in uint32?
	convertedProtocol := uint16(incomeIPVSService.Protocol)

	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)

	if err := gs.facade.AddIPVSApplicationServersForService(
		incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logger.Error().Err(err).Msg("failed to add real for vip")
		return &transport.EmptyGetIPVSData{}, err
	}

	// TODO not sure what log record can be useful here
	//gs.facade.Logger.Info().Msg("1")
	return &transport.EmptyGetIPVSData{}, nil
}

func (gs *GrpcServer) RemoveIPVSService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
	//gs.facade.Logger.Info().
	//	Uint32("protocol", incomeIPVSService.Protocol).
	//	Str("vip", incomeIPVSService.Vip).
	//	Uint32("port", incomeIPVSService.Port).
	//	Msg("removing vip")

	// FIXME why it comes in uint32?
	convertedPort := uint16(incomeIPVSService.Port)
	// FIXME why it comes in uint32?
	convertedProtocol := uint16(incomeIPVSService.Protocol)

	if err := gs.facade.RemoveIPVSService(
		incomeIPVSService.Vip,
		convertedPort,
		convertedProtocol,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logger.Error().Err(err).
			Uint32("protocol", incomeIPVSService.Protocol).
			Str("vip", incomeIPVSService.Vip).
			Uint32("port", incomeIPVSService.Port).
			Msg("failed to remove vip")
		return &transport.EmptyGetIPVSData{}, err
	}

	return &transport.EmptyGetIPVSData{}, nil
}

func (gs *GrpcServer) RemoveIPVSApplicationServersFromService(ctx context.Context, incomeIPVSService *transport.PbGetIPVSServiceData) (*transport.EmptyGetIPVSData, error) {
	//gs.facade.Logger.Info().
	//	Uint32("protocol", incomeIPVSService.Protocol).
	//	Str("vip", incomeIPVSService.Vip).
	//	Uint32("port", incomeIPVSService.Port).
	//	Msg("removing real from vip") // FIXME no source to get exact reals

	// FIXME why it comes in uint32?
	convertedPort := uint16(incomeIPVSService.Port)
	// FIXME why it comes in uint32?
	convertedProtocol := uint16(incomeIPVSService.Protocol)
	// FIXME why?
	convertedApplicationServers := convertPbApplicationServersToInternal(incomeIPVSService.ApplicationServers)

	if err := gs.facade.RemoveIPVSApplicationServersFromService(
		incomeIPVSService.Vip,
		convertedPort,
		incomeIPVSService.RoutingType,
		incomeIPVSService.BalanceType,
		convertedProtocol,
		convertedApplicationServers,
		incomeIPVSService.Id,
	); err != nil {
		gs.facade.Logger.Error().Err(err).Msg("failed to remove real from vip")
		return &transport.EmptyGetIPVSData{}, err
	}

	return &transport.EmptyGetIPVSData{}, nil
}

func (gs *GrpcServer) GetIPVSRuntime(ctx context.Context, emptyIPVSService *transport.EmptyGetIPVSData) (*transport.PbGetIPVSRawServicesData, error) {
	//gs.facade.Logger.Trace().Msg("getting ipvs runtime")
	ipvsRuntime, err := gs.facade.GetIPVSRuntime(emptyIPVSService.Id)
	if err != nil {
		gs.facade.Logger.Error().
			Err(err).
			Msg("failed to get ipvs runtime")
		return nil, err
	}

	pbIPVSServices := convertInternalServicesToPbServices(ipvsRuntime, emptyIPVSService.Id)

	//gs.facade.Logger.Info().Msg("successfully retrieved ipvs runtime")

	return pbIPVSServices, nil
}

func (gs *GrpcServer) StartServer() error {
	if err := gs.cleanup(); err != nil {
		gs.facade.Logger.Error().Err(err).Msg("failed to cleanup socket info")
		return err
	}

	lis, err := net.Listen("unix", gs.addr)
	//lis, err := net.Listen("tcp4", ":9000")
	if err != nil {
		gs.facade.Logger.Error().Err(err).Msg("failed to listen")
		return err
	}
	gs.grpcSrv = grpc.NewServer()
	transport.RegisterIPVSGetWorkerServer(gs.grpcSrv, gs)
	go gs.serve(lis)
	return nil
}

func (gs *GrpcServer) serve(lis net.Listener) {
	if err := gs.grpcSrv.Serve(lis); err != nil {
		gs.facade.Logger.Fatal().Err(err).Msg("failed to serve")
	}
}

func (gs *GrpcServer) CloseServer() {
	gs.grpcSrv.Stop()
	if err := gs.cleanup(); err != nil {
		gs.facade.Logger.Error().Err(err).Msg("failed to cleanup")
	}
}

func (gs *GrpcServer) cleanup() error {
	if _, err := os.Stat(gs.addr); err == nil {
		if err := os.RemoveAll(gs.addr); err != nil {
			return err
		}
	}
	return nil
}

func convertPbApplicationServersToInternal(pbApplicationServers map[string]uint32) map[string]uint16 {
	internalApplicationServers := make(map[string]uint16, len(pbApplicationServers))
	for k, v := range pbApplicationServers {
		internalApplicationServers[k] = uint16(v)
	}
	return internalApplicationServers
}

func convertInternalToPbApplicationServers(internalApplicationServers map[string]uint16) map[string]uint32 {
	pbApplicationServers := make(map[string]uint32, len(internalApplicationServers))
	for k, v := range internalApplicationServers {
		pbApplicationServers[k] = uint32(v)
	}
	return pbApplicationServers
}

func convertInternalServicesToPbServices(internalServices map[string]map[string]uint16, id string) *transport.PbGetIPVSRawServicesData {
	pbServices := make(map[string]*transport.PbGetRawIPVSServiceData, len(internalServices))
	for k, v := range internalServices {
		applicationServers := &transport.PbGetRawIPVSServiceData{
			RawServiceData: convertInternalToPbApplicationServers(v),
		}
		pbServices[k] = applicationServers
	}

	return &transport.PbGetIPVSRawServicesData{
		RawServicesData: pbServices,
		Id:              id,
	}
}
