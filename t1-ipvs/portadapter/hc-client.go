package portadapter

import (
	"context"
	"fmt"
	"net"
	"time"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type HealthcheckWorkerEntity struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	// conn          *grpc.ClientConn
	// hcWokerClient transport.SendRuntimeClient
	logging *logrus.Logger
}

func NewHealthcheckWorkerEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *HealthcheckWorkerEntity {
	return &HealthcheckWorkerEntity{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (healthcheckWorker *HealthcheckWorkerEntity) SendIPVSRuntime(runtimeConfig map[string]map[string]uint16,
	id string) error {
	withContextDialer := makeDialer(healthcheckWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	healthcheckClient := transport.NewSendIPVSRuntimeClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRuntime := convertInternalServicesToPbServices(runtimeConfig, id)
	_, err = healthcheckClient.SendIPVSRuntime(sendCtx, pbRuntime)
	return err
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

func convertInternalToPbApplicationServers(internalApplicationServers map[string]uint16) map[string]uint32 {
	pbApplicationServers := make(map[string]uint32, len(internalApplicationServers))
	for k, v := range internalApplicationServers {
		pbApplicationServers[k] = uint32(v)
	}
	return pbApplicationServers
}

func convertInternalServicesToPbServices(internalServices map[string]map[string]uint16, id string) *transport.PbSendIPVSRawServicesData {
	pbServices := make(map[string]*transport.PbSendRawIPVSServiceData, len(internalServices))
	for k, v := range internalServices {
		applicationServers := &transport.PbSendRawIPVSServiceData{
			RawServiceData: convertInternalToPbApplicationServers(v),
		}
		pbServices[k] = applicationServers
	}

	return &transport.PbSendIPVSRawServicesData{
		RawServicesData: pbServices,
		Id:              id,
	}
}
