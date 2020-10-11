package portadapter

import (
	context "context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type IPVSWorkerEntity struct {
	address         string
	grpcTimeout     time.Duration // TODO: somehow use tickers?
	conn            *grpc.ClientConn
	ipvsWokerClient IPVSWokerClient
	logging         *logrus.Logger
}

func NewIPVSWorkerEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *IPVSWorkerEntity {
	return &IPVSWorkerEntity{address: address, logging: logging}
}

func (ipvsWorker *IPVSWorkerEntity) initGRPC() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ipvsWorker.conn, err = grpc.DialContext(ctx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	ipvsWorker.ipvsWokerClient = NewIPVSWokerClient(ipvsWorker.conn)

	return nil
}

func convertMapToPbmap(incomeMap map[string]uint16) map[string]uint32 {
	convertedMap := make(map[string]uint32, len(incomeMap))
	for k, v := range incomeMap {
		convertedMap[k] = uint32(v)
	}
	return convertedMap
}

func (ipvsWorker *IPVSWorkerEntity) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	convertedMap := convertMapToPbmap(applicationServers)
	pbServiceInfo := &PbIPVSServices{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        routingType,
		BalanceType:        balanceType,
		Protocol:           uint32(protocol),
		ApplicationServers: convertedMap,
		Id:                 id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.ipvsWokerClient.NewIPVSService(ctx, pbServiceInfo)
	if err != nil {
		return fmt.Errorf("can't add new service to ipvs: %v", err)
	}
	return nil
}

func (ipvsWorker *IPVSWorkerEntity) AddIPVSApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	convertedMap := convertMapToPbmap(applicationServers)
	pbServiceInfo := &PbIPVSServices{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        routingType,
		BalanceType:        balanceType,
		Protocol:           uint32(protocol),
		ApplicationServers: convertedMap,
		Id:                 id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.ipvsWokerClient.AddIPVSApplicationServersForService(ctx, pbServiceInfo)
	if err != nil {
		return fmt.Errorf("can't add new service to ipvs: %v", err)
	}
	return nil
}

func (ipvsWorker *IPVSWorkerEntity) RemoveIPVSService(vip string,
	port uint16,
	protocol uint16,
	id string) error {
	pbServiceInfo := &PbIPVSServices{
		Vip:  vip,
		Port: uint32(port),
		Id:   id,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.ipvsWokerClient.RemoveIPVSService(ctx, pbServiceInfo)
	if err != nil {
		return fmt.Errorf("can't add new service to ipvs: %v", err)
	}
	return nil
}

func (ipvsWorker *IPVSWorkerEntity) RemoveIPVSApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	convertedMap := convertMapToPbmap(applicationServers)
	pbServiceInfo := &PbIPVSServices{
		Vip:                vip,
		Port:               uint32(port),
		RoutingType:        routingType,
		BalanceType:        balanceType,
		Protocol:           uint32(protocol),
		ApplicationServers: convertedMap,
		Id:                 id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.ipvsWokerClient.RemoveIPVSApplicationServersFromService(ctx, pbServiceInfo)
	if err != nil {
		return fmt.Errorf("can't add new service to ipvs: %v", err)
	}
	return nil
}

func (ipvsWorker *IPVSWorkerEntity) IsIPVSApplicationServerInService(serviceIP string,
	servicePort uint16,
	oneApplicationServerMap map[string]uint16,
	id string) (bool, error) {
	convertedMap := convertMapToPbmap(oneApplicationServerMap)
	pbServiceInfo := &PbIPVSServices{
		Vip:                serviceIP,
		Port:               uint32(servicePort),
		ApplicationServers: convertedMap,
		Id:                 id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	isIn, err := ipvsWorker.ipvsWokerClient.IsIPVSApplicationServerInService(ctx, pbServiceInfo)
	if err != nil {
		return false, fmt.Errorf("can't add new service to ipvs: %v", err)
	}

	return isIn.IsIn, nil
}

func (ipvsWorker *IPVSWorkerEntity) IPVSFlush() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.ipvsWokerClient.IPVSFlush(ctx, &EmptyPbService{})
	return err
}

func (ipvsWorker *IPVSWorkerEntity) ConnectToIPVS() error {
	return ipvsWorker.initGRPC()
}

func (ipvsWorker *IPVSWorkerEntity) DisconnectFromIPVS() {
	if err := ipvsWorker.conn.Close(); err != nil {
		ipvsWorker.logging.Errorf("close grpc connection to hc error: %v", err)
	}
}
