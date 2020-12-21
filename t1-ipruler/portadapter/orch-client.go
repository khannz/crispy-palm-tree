package portadapter

import (
	"context"
	"fmt"
	"net"
	"time"

	transport "github.com/khannz/crispy-palm-tree/t1-ipruler/grpc-orch"
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

func (orchestratorWorker *OrchestratorWorkerEntity) SendRouteRuntimeConfig(runtimeConfig map[int]struct{},
	id string) error {
	convertedRuntimeConfig := convertMapToPbMap(runtimeConfig)

	withContextDialer := makeDialer(orchestratorWorker.address, 2*time.Second)
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, orchestratorWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	healthcheckClient := transport.NewSendDummyRuntimeClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRuntimeConfig := &transport.SendIpRuleRuntimeData{
		Fwmarks: convertedRuntimeConfig,
		Id:      id,
	}

	_, err = healthcheckClient.SendIpRuleRuntime(sendCtx, pbRuntimeConfig)
	return err
}

func convertMapToPbMap(currentConfig map[int]struct{}) map[int64]*transport.EmptySendIpRuleData {
	convertedCurrentConfig := make(map[int64]*transport.EmptySendIpRuleData, len(currentConfig))
	for cc := range currentConfig {
		convertedCurrentConfig[int64(cc)] = &transport.EmptySendIpRuleData{}
	}
	return convertedCurrentConfig
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
