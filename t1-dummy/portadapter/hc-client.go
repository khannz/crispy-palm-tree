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

func (healthcheckWorker *HealthcheckWorkerEntity) SendDummyRuntimeConfig(runtimeConfig map[string]struct{},
	id string) error {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckWorker.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	defer conn.Close()
	healthcheckClient := transport.NewSendDummyRuntimeClient(conn)

	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRuntimeConfig := convertRuntimeConfigToPbRuntimeConfig(runtimeConfig, id)
	_, err = healthcheckClient.SendDummyRuntime(sendCtx, pbRuntimeConfig)
	return err
}

func convertRuntimeConfigToPbRuntimeConfig(runtimeConfig map[string]struct{}, id string) *transport.SendDummyRuntimeData {
	ed := &transport.EmptySendDummyData{}
	pbMap := make(map[string]*transport.EmptySendDummyData)
	for k := range runtimeConfig {
		pbMap[k] = ed
	}

	return &transport.SendDummyRuntimeData{
		Services: pbMap,
		Id:       id,
	}
}
