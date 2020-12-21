package portadapter

import (
	"context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-ipruler"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type IpRuleWorker struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	logging     *logrus.Logger
}

func NewIpRuleWorker(address string, grpcTimeout time.Duration, logging *logrus.Logger) *IpRuleWorker {
	return &IpRuleWorker{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (ipRuleWorker *IpRuleWorker) AddIPRule(hcTunDestIP string, id string) error {
	withContextDialer := makeDialer(ipRuleWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipRuleWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewIPRulerGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbAddRoute := &transport.IpData{
		TunDestIP: hcTunDestIP,
		Id:        id,
	}
	_, err = routeClient.AddToIPRuler(sendCtx, pbAddRoute)
	return err
}

func (ipRuleWorker *IpRuleWorker) RemoveIPRule(hcTunDestIP string, id string) error {
	withContextDialer := makeDialer(ipRuleWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipRuleWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewIPRulerGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRemoveRoute := &transport.IpData{
		TunDestIP: hcTunDestIP,
		Id:        id,
	}
	_, err = routeClient.RemoveFromIPRuler(sendCtx, pbRemoveRoute)
	return err
}

// GetIPRulerRuntime ...
func (ipRuleWorker *IpRuleWorker) GetIPRulerRuntime(id string) (map[int]struct{}, error) {
	withContextDialer := makeDialer(ipRuleWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipRuleWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return nil, fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewIPRulerGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbGetRouteRuntimeConfigRequest := &transport.EmptyGetIPRulerData{Id: id}
	pbRouteRuntimeConfig, err := routeClient.GetIPRulerRuntime(sendCtx, pbGetRouteRuntimeConfigRequest)
	convertedIpRuls := convertIpRuls(pbRouteRuntimeConfig.Fwmarks)
	return convertedIpRuls, err
}

func convertIpRuls(pbIpRules map[int64]*transport.EmptyGetIPRulerData) map[int]struct{} {
	convertedMap := make(map[int]struct{}, len(pbIpRules))
	for k := range pbIpRules {
		convertedMap[int(k)] = struct{}{} // FIXME: may be broken
	}
	return convertedMap
}
