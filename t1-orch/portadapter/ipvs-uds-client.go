package portadapter

import (
	"context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-ipvs"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type IpvsWorker struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	logging     *logrus.Logger
}

func NewIpvsWorker(address string, grpcTimeout time.Duration, logging *logrus.Logger) *IpvsWorker {
	return &IpvsWorker{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (ipvsWorker *IpvsWorker) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	id string) error {
	withContextDialer := makeDialer(ipvsWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	ipvsClient := transport.NewIPVSGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbNewIPVSService := &transport.PbGetIPVSServiceData{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        routingType,
		BalanceType:        balanceType,
		Protocol:           uint32(protocol),
		ApplicationServers: nil,
		Id:                 id,
	}
	_, err = ipvsClient.NewIPVSService(sendCtx, pbNewIPVSService)
	return err
}

func (ipvsWorker *IpvsWorker) AddIPVSApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	convertedApplicationServers := convertApplicationServers(applicationServers)
	withContextDialer := makeDialer(ipvsWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	ipvsClient := transport.NewIPVSGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbAddIPVSApplicationServersForService := &transport.PbGetIPVSServiceData{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        routingType,
		BalanceType:        balanceType,
		Protocol:           uint32(protocol),
		ApplicationServers: convertedApplicationServers,
		Id:                 id,
	}
	_, err = ipvsClient.AddIPVSApplicationServersForService(sendCtx, pbAddIPVSApplicationServersForService)
	return err
}

func (ipvsWorker *IpvsWorker) RemoveIPVSService(vip string,
	port uint16,
	protocol uint16,
	id string) error {
	withContextDialer := makeDialer(ipvsWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	ipvsClient := transport.NewIPVSGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRemoveIPVSService := &transport.PbGetIPVSServiceData{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        0,
		BalanceType:        "",
		Protocol:           uint32(protocol),
		ApplicationServers: nil,
		Id:                 id,
	}
	_, err = ipvsClient.RemoveIPVSService(sendCtx, pbRemoveIPVSService)
	return err
}

func (ipvsWorker *IpvsWorker) RemoveIPVSApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	convertedApplicationServers := convertApplicationServers(applicationServers)
	withContextDialer := makeDialer(ipvsWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	ipvsClient := transport.NewIPVSGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRemoveIPVSApplicationServersFromService := &transport.PbGetIPVSServiceData{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        routingType,
		BalanceType:        balanceType,
		Protocol:           uint32(protocol),
		ApplicationServers: convertedApplicationServers,
		Id:                 id,
	}
	_, err = ipvsClient.RemoveIPVSApplicationServersFromService(sendCtx, pbRemoveIPVSApplicationServersFromService)
	return err
}

func (ipvsWorker *IpvsWorker) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	withContextDialer := makeDialer(ipvsWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return nil, fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	ipvsClient := transport.NewIPVSGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbGetIPVSRuntimeRequest := &transport.EmptyGetIPVSData{Id: id}
	pbIPVSRuntime, err := ipvsClient.GetIPVSRuntime(sendCtx, pbGetIPVSRuntimeRequest)
	if err != nil {
		return nil, err
	}
	convertedIPVSRuntime := convertPbIPVSRuntime(pbIPVSRuntime)
	return convertedIPVSRuntime, err
}

func convertApplicationServers(applicationServers map[string]uint16) map[string]uint32 {
	newMap := make(map[string]uint32, len(applicationServers))
	for k, v := range applicationServers {
		newMap[k] = uint32(v)
	}
	return newMap
}

func convertPbIPVSRuntime(pbGetIPVSRuntime *transport.PbGetIPVSRawServicesData) map[string]map[string]uint16 {
	servicesMap := make(map[string]map[string]uint16, len(pbGetIPVSRuntime.GetRawServicesData()))
	for sk, sv := range pbGetIPVSRuntime.RawServicesData {
		serviceMap := make(map[string]uint16, len(sv.GetRawServiceData()))
		for k, v := range sv.GetRawServiceData() {
			serviceMap[k] = uint16(v)
		}
		servicesMap[sk] = serviceMap
	}
	return nil
}
