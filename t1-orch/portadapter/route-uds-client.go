package portadapter

import (
	"context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-route"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type RouteWorker struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	logging     *logrus.Logger
}

func NewRouteWorker(address string, grpcTimeout time.Duration, logging *logrus.Logger) *RouteWorker {
	return &RouteWorker{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (routeWorker *RouteWorker) AddRoute(hcDestIP string, hcTunDestIP string, id string) error {
	withContextDialer := makeDialer(routeWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, routeWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewRouteGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbAddRoute := &transport.RouteData{
		HcDestIP:    hcDestIP,
		HcTunDestIP: hcTunDestIP,
		Id:          id,
	}
	_, err = routeClient.AddRoute(sendCtx, pbAddRoute)
	return err
}

func (routeWorker *RouteWorker) RemoveRoute(hcDestIP string, hcTunDestIP string, id string) error {
	withContextDialer := makeDialer(routeWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, routeWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewRouteGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRemoveRoute := &transport.RouteData{
		HcDestIP:    hcDestIP,
		HcTunDestIP: hcTunDestIP,
		Id:          id,
	}
	_, err = routeClient.RemoveRoute(sendCtx, pbRemoveRoute)
	return err
}

func (routeWorker *RouteWorker) GetRouteRuntimeConfig(id string) ([]string, error) {
	withContextDialer := makeDialer(routeWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, routeWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return nil, fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	routeClient := transport.NewRouteGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbGetRouteRuntimeConfigRequest := &transport.EmptyRouteData{Id: id}
	pbRouteRuntimeConfig, err := routeClient.GetRouteRuntimeConfig(sendCtx, pbGetRouteRuntimeConfigRequest)
	return pbRouteRuntimeConfig.RouteData, err
}
