package portadapter

import (
	"context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-tunnel"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type TunnelWorker struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	logging     *logrus.Logger
}

func NewTunnelWorker(address string, grpcTimeout time.Duration, logging *logrus.Logger) *TunnelWorker {
	return &TunnelWorker{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (tunnelWorker *TunnelWorker) AddTunnel(hcTunDestIP string, id string) error {
	withContextDialer := makeDialer(tunnelWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, tunnelWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewTunnelGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbAddTunnel := &transport.TunnelData{
		HcTunDestIP: hcTunDestIP,
		Id:          id,
	}
	_, err = routeClient.AddTunnel(sendCtx, pbAddTunnel)
	return err
}

func (tunnelWorker *TunnelWorker) RemoveTunnel(hcTunDestIP string, id string) error {
	withContextDialer := makeDialer(tunnelWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, tunnelWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewTunnelGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRemoveTunnel := &transport.TunnelData{
		HcTunDestIP: hcTunDestIP,
		Id:          id,
	}
	_, err = routeClient.RemoveTunnel(sendCtx, pbRemoveTunnel)
	return err
}

func (tunnelWorker *TunnelWorker) GetTunnelRuntime(id string) (map[string]struct{}, error) {
	withContextDialer := makeDialer(tunnelWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, tunnelWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return nil, fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewTunnelGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbGetTunnelRuntimeConfigRequest := &transport.EmptyTunnelData{Id: id}
	pbTunnelRuntimeConfig, err := routeClient.GetTunnelRuntime(sendCtx, pbGetTunnelRuntimeConfigRequest)
	convertedTunnels := convertTunnels(pbTunnelRuntimeConfig.Tunnels)
	return convertedTunnels, err
}

func convertTunnels(pbTunnelMap map[string]int32) map[string]struct{} {
	convertedMap := make(map[string]struct{}, len(pbTunnelMap))
	for k := range pbTunnelMap {
		convertedMap[k] = struct{}{}
	}
	return convertedMap
}
