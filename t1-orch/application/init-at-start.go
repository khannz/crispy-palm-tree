package application

import (
	"fmt"
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/usecase"
)

// InitConfigAtStart ...
func (t1OrchFacade *T1OrchFacade) InitConfigAtStart(agentID, id string) error {
	currentServices, err := getCurrentConfigFromKVStorage(agentID)
	if err != nil {
		return fmt.Errorf("can't init config: get current config fail: %v", err)
	}

	newNewServiceEntity := usecase.NewNewServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, currentService := range currentServices {
		enrichKVServiceDataToDomainServiceInfo(currentService) // add data logic data fields
		if err := t1OrchFacade.MemoryWorker.AddService(currentService); err != nil {
			return err
		}
		if err := newNewServiceEntity.NewService(currentService, id); err != nil {
			return err
		}
	}
	return nil
}

func getCurrentConfigFromKVStorage(agengID string) ([]*domain.ServiceInfo, error) {

	return mockServices(), nil
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
