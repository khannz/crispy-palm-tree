package application

import (
	"encoding/json"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

const consulWorkerName = "consul worker"

// ConsulWorker ...
type ConsulWorker struct {
	facade                          *T1OrchFacade
	client                          *consulapi.Client
	kvClient                        *consulapi.KV
	subscribePath                   string
	appServersPath                  string
	serviceManifest                 string
	needToRemoveServicesIfDataEmpty bool
	jobChan                         chan map[string]*domain.ServiceInfo
	logging                         *logrus.Logger
}

// NewConsulWorker ...
func NewConsulWorker(facade *T1OrchFacade,
	consulAddress,
	consulSubscribePath,
	consulAppServersPath,
	serviceManifest string,
	logging *logrus.Logger) (*ConsulWorker, error) {
	clientAPI := consulapi.DefaultConfig()
	clientAPI.Address = consulAddress
	client, err := consulapi.NewClient(clientAPI)
	if err != nil {
		return nil, err
	}
	return &ConsulWorker{
		facade:                          facade,
		client:                          client,
		kvClient:                        client.KV(),
		subscribePath:                   consulSubscribePath,
		appServersPath:                  consulAppServersPath,
		serviceManifest:                 serviceManifest,
		needToRemoveServicesIfDataEmpty: true,
		jobChan:                         make(chan map[string]*domain.ServiceInfo),
		logging:                         logging,
	}, nil
}

func (consulWorker *ConsulWorker) ConsulConfigWatch() {
	defer close(consulWorker.jobChan)
	currentIndex := uint64(0)
	queryOptions := &consulapi.QueryOptions{WaitIndex: currentIndex}
	for {
		balancingServices, meta, err := consulWorker.kvClient.Keys(consulWorker.subscribePath, "/", queryOptions)
		if err != nil {
			consulWorker.logging.WithFields(logrus.Fields{
				"entity": consulWorkerName,
			}).Errorf("can't get balancing services: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if balancingServices == nil || meta == nil || len(balancingServices) <= 1 {
			if consulWorker.needToRemoveServicesIfDataEmpty {
				consulWorker.logging.WithFields(logrus.Fields{
					"entity": consulWorkerName,
				}).Infof("balancing services not found. started deleting existing services: %v", err)
				consulWorker.jobChan <- nil
				consulWorker.needToRemoveServicesIfDataEmpty = false
			}
			time.Sleep(5 * time.Second)
			continue
		} else if currentIndex == meta.LastIndex {
			// ~ every 300-330 sec indexes autocheck
			consulWorker.logging.WithFields(logrus.Fields{
				"entity": consulWorkerName,
			}).Tracef("metaindex not change: %v", currentIndex)
			time.Sleep(1 * time.Second)
			continue
		}

		consulWorker.needToRemoveServicesIfDataEmpty = true

		currentIndex = meta.LastIndex

		servicesInfo, err := consulWorker.formUpdateServicesInfo(balancingServices)
		if err != nil {
			consulWorker.logging.WithFields(logrus.Fields{
				"entity": consulWorkerName,
			}).Errorf("can't form update services info: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		consulWorker.jobChan <- servicesInfo
	}
}

func (consulWorker *ConsulWorker) formUpdateServicesInfo(balancingServices []string) (map[string]*domain.ServiceInfo, error) {
	balancingServicesTransportArray := make([]*ServiceTransport, 0, len(balancingServices)-1)
	for _, bsPath := range balancingServices {
		if bsPath == consulWorker.subscribePath {
			continue
		}

		applicationServersPaths, _, err := consulWorker.kvClient.Keys(bsPath+consulWorker.appServersPath, "/", nil)
		if err != nil {
			return nil, fmt.Errorf("can't get application servers paths: %v", err)
		}

		if len(applicationServersPaths) <= 1 {
			return nil, fmt.Errorf("application servers not found for service %v", bsPath)
		}

		applicationServersTransportArray := make([]*ApplicationServerTransport, 0, len(applicationServersPaths)-1)
		for _, applicationServersPath := range applicationServersPaths {
			if applicationServersPath == bsPath+consulWorker.appServersPath {
				continue
			}
			applicationServerPair, _, err := consulWorker.kvClient.Get(applicationServersPath, nil)
			if err != nil {
				return nil, fmt.Errorf("can't get application servers pair: %v", err)
			}
			applicationServerTransport := &ApplicationServerTransport{}
			if err := json.Unmarshal(applicationServerPair.Value, applicationServerTransport); err != nil {
				return nil, fmt.Errorf("can't get unmarshall application server transport: %v", err)
			}
			applicationServersTransportArray = append(applicationServersTransportArray, applicationServerTransport)
		}
		serviceManifestPair, _, err := consulWorker.kvClient.Get(bsPath+consulWorker.serviceManifest, nil)
		if err != nil || serviceManifestPair == nil {
			return nil, fmt.Errorf("can't get service manifest pair: %v", err)
		}
		balancingServiceTransport := &ServiceTransport{}
		if err := json.Unmarshal(serviceManifestPair.Value, balancingServiceTransport); err != nil {
			return nil, fmt.Errorf("can't get unmarshall service transport: %v", err)
		}
		balancingServiceTransport.ApplicationServersTransport = applicationServersTransportArray
		balancingServicesTransportArray = append(balancingServicesTransportArray, balancingServiceTransport)
	}

	servicesInfo, err := convertBalancingServicesTransportArrayToDomainModel(balancingServicesTransportArray)
	if err != nil {
		return nil, fmt.Errorf("can't convert services transport to domain model: %v", err)
	}
	return servicesInfo, nil
}

func (consulWorker *ConsulWorker) JobWorker() {
	for servicesInfo := range consulWorker.jobChan {
		if servicesInfo != nil {
			if err := consulWorker.facade.ApplyNewConfig(servicesInfo); err != nil {
				consulWorker.logging.WithFields(logrus.Fields{
					"entity": consulWorkerName,
				}).Errorf("config update error: %v", err)
			}
		} else {
			if err := consulWorker.facade.RemoveAllConfig(); err != nil {
				consulWorker.logging.WithFields(logrus.Fields{
					"entity": consulWorkerName,
				}).Errorf("config update error: %v", err)
			}
		}
	}
}
