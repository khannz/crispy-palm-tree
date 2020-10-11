package portadapter

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type IPVSSenderEntity struct {
	address         string
	grpcTimeout     time.Duration // TODO: somehow use tickers?
	conn            *grpc.ClientConn
	ipvsWokerClient IPVSWokerClient
	logging         *logrus.Logger
}

func NewIPVSSenderEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *IPVSSenderEntity {
	return &IPVSSenderEntity{address: address, logging: logging}
}

func (ipvsSender *IPVSSenderEntity) initGRPC() error {
	var err error
	ipvsSender.conn, err = grpc.Dial(ipvsSender.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	ipvsSender.ipvsWokerClient = NewIPVSWokerClient(ipvsSender.conn)

	return nil
}

func (ipvsSender *IPVSSenderEntity) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	// ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	// defer cancel()
	return nil
}

func (ipvsSender *IPVSSenderEntity) AddIPVSApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {

	return nil
}

func (ipvsSender *IPVSSenderEntity) RemoveIPVSService(vip string,
	port uint16,
	protocol uint16,
	id string) error {

	return nil
}

func (ipvsSender *IPVSSenderEntity) RemoveIPVSApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {

	return nil
}

func (ipvsSender *IPVSSenderEntity) IsIPVSApplicationServerInService(serviceIP string,
	servicePort uint16,
	oneApplicationServerMap map[string]uint16,
	id string) (bool, error) {

	return false, nil
}

func (ipvsSender *IPVSSenderEntity) IPVSFlush() error {

	return nil
}

func (ipvsSender *IPVSSenderEntity) ConnectToHealtchecks() error {
	return ipvsSender.initGRPC()
}
func (ipvsSender *IPVSSenderEntity) DisconnectFromHealtchecks() {
	if err := ipvsSender.conn.Close(); err != nil {
		ipvsSender.logging.Errorf("close grpc connection to hc error: %v", err)
	}
}
