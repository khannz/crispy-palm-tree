package portadapter

import (
	"context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-dummy/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type HealthcheckWorkerEntity struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	// conn          *grpc.ClientConn
	// hcWokerClient transport.GetRuntimeClient
	logging *logrus.Logger
}

func NewHealthcheckWorkerEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *HealthcheckWorkerEntity {
	return &HealthcheckWorkerEntity{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (healthcheckWorker *HealthcheckWorkerEntity) SendRuntimeConfig(runtimeConfig map[string]struct{},
	id string) error {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckWorker.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	defer conn.Close()
	healthcheckClient := transport.NewSendRuntimeClient(conn)

	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRuntimeConfig := convertRuntimeConfigToPbRuntimeConfig(runtimeConfig, id)
	_, err = healthcheckClient.DummySendRuntime(sendCtx, pbRuntimeConfig)
	return err
}

func convertRuntimeConfigToPbRuntimeConfig(runtimeConfig map[string]struct{}, id string) *transport.DummySendRuntimeData {
	ed := &transport.EmptySendData{}
	pbMap := make(map[string]*transport.EmptySendData)
	for k := range runtimeConfig {
		pbMap[k] = ed
	}

	return &transport.DummySendRuntimeData{
		Services: pbMap,
		Id:       id,
	}
}
