package portadapter

import (
	"context"
	"fmt"
	"time"

	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-dummy"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DummyWorker struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	logging     *logrus.Logger
}

func NewDummyWorker(address string, grpcTimeout time.Duration, logging *logrus.Logger) *DummyWorker {
	return &DummyWorker{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (dummyWorker *DummyWorker) AddToDummy(ip string, id string) error {
	withContextDialer := makeDialer(dummyWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, dummyWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	dummyClient := transport.NewDummyGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbAddToDummy := &transport.IpData{
		Ip: ip,
		Id: id,
	}
	_, err = dummyClient.AddToDummy(sendCtx, pbAddToDummy)
	return err
}

func (dummyWorker *DummyWorker) RemoveFromDummy(ip string, id string) error {
	withContextDialer := makeDialer(dummyWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, dummyWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	dummyClient := transport.NewDummyGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbRemoveFromDummy := &transport.IpData{
		Ip: ip,
		Id: id,
	}
	_, err = dummyClient.RemoveFromDummy(sendCtx, pbRemoveFromDummy)
	return err
}

func (dummyWorker *DummyWorker) GetDummyRuntimeConfig(id string) (map[string]struct{}, error) {
	// not implemented
	withContextDialer := makeDialer(dummyWorker.address, 2*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, dummyWorker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return nil, fmt.Errorf("can't connect to grpc uds server: %v", err)
	}
	defer conn.Close()

	dummyClient := transport.NewDummyGetWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sendCancel()

	pbGetDummy := &transport.EmptyGetDummyData{Id: id}
	pbDummyRuntimeData, err := dummyClient.GetDummyRuntime(sendCtx, pbGetDummy)
	if err != nil {
		return nil, err
	}
	dummyRuntimeData := convertPbDummyMap(pbDummyRuntimeData)
	return dummyRuntimeData, err
}

func convertPbDummyMap(pb *transport.GetDummyRuntimeData) map[string]struct{} {
	dm := make(map[string]struct{}, len(pb.Services))
	for s := range pb.Services {
		dm[s] = struct{}{}
	}
	return dm
}
