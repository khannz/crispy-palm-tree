package application

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
)

const etcdConfig = "etcd config"

type EtcdWorker struct {
	facade     *T1OrchFacade
	EtcdClient *clientv3.Client
	agentID    string
}

func NewEtcdWorker(facade *T1OrchFacade,
	endpoints []string,
	dialTimeout time.Duration,
	agentID string) (*EtcdWorker, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdWorker{
		facade:     facade,
		EtcdClient: etcdClient,
		agentID:    agentID,
	}, nil
}

func (etcdWorker *EtcdWorker) EtcdConfigWatch() {
	getResp, err := etcdWorker.EtcdClient.Get(context.Background(), etcdWorker.agentID)
	if err != nil {
		etcdWorker.facade.Logging.WithFields(logrus.Fields{
			"entity":   etcdConfig,
			"event id": etcdWorker.agentID,
		}).Warnf("can't read new config from etcd: %v", err)
	}

	if len(getResp.OpResponse().Get().Kvs) != 0 {
		kvs := string(getResp.OpResponse().Get().Kvs[0].Value)
		etcdWorker.facade.Logging.WithFields(logrus.Fields{
			"entity":   etcdConfig,
			"event id": etcdWorker.agentID,
		}).Warnf("get init config from etcd: %v", kvs)
		etcdWorker.facade.ApplyNewConfig()
	}

	watchChan := etcdWorker.EtcdClient.Watch(context.Background(), etcdWorker.agentID)
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			etcdWorker.facade.Logging.WithFields(logrus.Fields{
				"entity":   etcdConfig,
				"event id": etcdWorker.agentID,
			}).Infof("got new config from etcd: %v", event.Kv.Value)
			etcdWorker.facade.ApplyNewConfig()
			// fmt.Printf("Event received! %s executed on %q with value %q\n", event.Type, event.Kv.Key, event.Kv.Value)
		}
	}
}
