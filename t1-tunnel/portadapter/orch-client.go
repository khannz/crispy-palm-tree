package portadapter

import (
	"context"
	"fmt"
	"net"
	"time"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/grpc-orch"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type OrchestratorWorkerEntity struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	// conn          *grpc.ClientConn
	// hcWokerClient transport.GetRuntimeClient
	logging *logrus.Logger
}

func NewOrchestratorWorkerEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *OrchestratorWorkerEntity {
	return &OrchestratorWorkerEntity{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (orchestratorWorker *OrchestratorWorkerEntity) SendTunnelRuntimeConfig(runtimeConfig map[string]struct{},
	id string) error {
	withContextDialer := makeDialer(orchestratorWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, orchestratorWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	healthcheckClient := transport.NewSendRuntimeClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	convertedRuntimeConfig := convertRuntimeConfig(runtimeConfig)

	pbRuntimeConfig := &transport.SendTunnelRuntimeData{
		Tunnels: convertedRuntimeConfig,
		Id:      id,
	}

	_, err = healthcheckClient.SendTunnelRuntime(sendCtx, pbRuntimeConfig)
	return err
}

func convertRuntimeConfig(runtimeConfig map[string]struct{}) map[string]int32 {
	convertedRuntimeConfig := make(map[string]int32, len(runtimeConfig))
	for k := range runtimeConfig {
		convertedRuntimeConfig[k] = 0
	}
	return convertedRuntimeConfig
}

func makeDialer(addr string, t time.Duration) func(ctx context.Context, addr string) (net.Conn, error) {
	f := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial("unix", addr)
	}

	return func(ctx context.Context, addr string) (net.Conn, error) {
		if deadline, ok := ctx.Deadline(); ok {
			return f(addr, time.Until(deadline))
		}
		return f(addr, 0)
	}
}
