package portadapter

import (
	context "context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/grpc-transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DummyWorkerEntity struct {
	address          string
	grpcTimeout      time.Duration // TODO: somehow use tickers?
	conn             *grpc.ClientConn
	dummyWokerClient transport.DummyWokerClient
	logging          *logrus.Logger
}

func NewDummyWorkerEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *DummyWorkerEntity {
	return &DummyWorkerEntity{address: address, logging: logging}
}

func (ipvsWorker *DummyWorkerEntity) initGRPC() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ipvsWorker.conn, err = grpc.DialContext(ctx, ipvsWorker.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	ipvsWorker.dummyWokerClient = transport.NewDummyWokerClient(ipvsWorker.conn)

	return nil
}

func (ipvsWorker *DummyWorkerEntity) AddToDummy(ip string, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.dummyWokerClient.AddToDummy(ctx, &transport.IpData{Ip: ip, Id: id})

	return err
}

func (ipvsWorker *DummyWorkerEntity) RemoveFromDummy(ip string, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ipvsWorker.dummyWokerClient.RemoveFromDummy(ctx, &transport.IpData{Ip: ip, Id: id})
	return err
}

func (ipvsWorker *DummyWorkerEntity) ConnectToDummy() error {
	return ipvsWorker.initGRPC()
}

func (ipvsWorker *DummyWorkerEntity) DisconnectFromDummy() {
	if err := ipvsWorker.conn.Close(); err != nil {
		ipvsWorker.logging.Errorf("close grpc connection to hc error: %v", err)
	}
}
