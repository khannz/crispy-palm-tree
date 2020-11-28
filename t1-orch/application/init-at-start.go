package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

func (etcdWorker *EtcdWorker) TmpEtcdPut() {
	mockServices := mockServices()

	out, err := json.Marshal(mockServices)
	if err != nil {
		panic(err)
	}

	// etcd.Put(context.Background(), *etcdWatchKey, time.Now().String())
	if _, err := etcdWorker.EtcdClient.Put(context.Background(), etcdWorker.agentID, string(out)); err != nil {
		fmt.Println("put to etcd error: ", err)
	}
	fmt.Println("populated " + etcdWorker.agentID + " with a value..")
}

func mockServices() []*domain.ServiceInfo {
	ds1a1 := &domain.ApplicationServer{
		Address:   "11.11.11.11:11",
		IP:        "11.11.11.11",
		Port:      "11",
		IsUp:      false,
		HCAddress: "11.11.11.11:8011",
	}
	ds1a2 := &domain.ApplicationServer{
		Address:   "12.12.12.12:12",
		IP:        "12.12.12.12",
		Port:      "12",
		IsUp:      false,
		HCAddress: "12.12.12.12:8012",
	}
	ds1am := map[string]*domain.ApplicationServer{"11.11.11.11:11": ds1a1, "12.12.12.12:12": ds1a2}
	ds1 := &domain.ServiceInfo{
		Address:               "1.1.1.1:80",
		IP:                    "1.1.1.1",
		Port:                  "80",
		IsUp:                  false,
		BalanceType:           "mhf",
		RoutingType:           "tunneling",
		Protocol:              "tcp",
		AlivedAppServersForUp: 1,
		HCType:                "tcp",
		HCRepeat:              2 * time.Second,
		HCTimeout:             1 * time.Second,
		HCRetriesForUP:        5,
		HCRetriesForDown:      2,
		ApplicationServers:    ds1am,
	}
	//
	ds2a1 := &domain.ApplicationServer{
		Address:   "21.21.21.21:21",
		IP:        "21.21.21.21",
		Port:      "21",
		IsUp:      false,
		HCAddress: "21.21.21.21:8021",
	}
	ds2a2 := &domain.ApplicationServer{
		Address:   "22.22.22.22:22",
		IP:        "22.22.22.22",
		Port:      "22",
		IsUp:      false,
		HCAddress: "22.22.22.22:8022",
	}
	ds2am := map[string]*domain.ApplicationServer{"11.11.11.11:11": ds2a1, "12.12.12.12:12": ds2a2}
	ds2 := &domain.ServiceInfo{
		Address:               "2.2.2.2:80",
		IP:                    "2.2.2.2",
		Port:                  "80",
		IsUp:                  false,
		BalanceType:           "mhf",
		RoutingType:           "tunneling",
		Protocol:              "tcp",
		AlivedAppServersForUp: 1,
		HCType:                "tcp",
		HCRepeat:              2 * time.Second,
		HCTimeout:             1 * time.Second,
		HCRetriesForUP:        5,
		HCRetriesForDown:      2,
		ApplicationServers:    ds2am,
	}

	return []*domain.ServiceInfo{ds1, ds2}
}
