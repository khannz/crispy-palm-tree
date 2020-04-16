package application

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/khannz/crispy-palm-tree/usecase"
	"github.com/sirupsen/logrus"
)

// BalancerFacade struct
type BalancerFacade struct {
	Locker            *domain.Locker
	VRRPConfigurator  domain.ServiceWorker
	CacheStorage      *portadapter.StorageEntity // so dirty
	PersistentStorage *portadapter.StorageEntity // so dirty
	TunnelConfig      domain.TunnelMaker
	UUIDgenerator     domain.UUIDgenerator
	Logging           *logrus.Logger
}

// NewBalancerFacade ...
func NewBalancerFacade(locker *domain.Locker,
	vrrpConfigurator domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		Locker:            locker,
		VRRPConfigurator:  vrrpConfigurator,
		CacheStorage:      cacheStorage,
		PersistentStorage: persistentStorage,
		TunnelConfig:      tunnelConfig,
		UUIDgenerator:     uuidGenerator,
		Logging:           logging,
	}
}

// CreateService ...
func (balancerFacade *BalancerFacade) CreateService(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	createServiceUUID string) error {
	newCreateServiceEntity := usecase.NewCreateServiceEntity(balancerFacade.Locker,
		balancerFacade.VRRPConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	serviceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, applicationServers)
	return newCreateServiceEntity.CreateService(serviceInfo, createServiceUUID)
}

// RemoveService ...
func (balancerFacade *BalancerFacade) RemoveService(serviceIP, servicePort string, newNWBRequestUUID string) error {
	removeService := usecase.NewRemoveServiceEntity(balancerFacade.Locker,
		balancerFacade.VRRPConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	serviceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, nil)
	return removeService.RemoveService(serviceInfo, newNWBRequestUUID)
}

// // GetNWBServices ...
// func (balancerFacade *BalancerFacade) GetNWBServices(getNWBServicesUUID string) ([]domain.ServiceInfo, error) {
// 	getNWBServices := usecase.NewGetNlbServices(balancerFacade.Logging)
// 	nwbServices, err := getNWBServices.GetAllNWBServices(getNWBServicesUUID)
// 	if err != nil {
// 		return nil, fmt.Errorf("can't get nwb services: %v", err)
// 	}
// 	return nwbServices, nil
// }

// AddApplicationServers ...
func (balancerFacade *BalancerFacade) AddApplicationServers(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	addApplicationServersRequestUUID string) (domain.ServiceInfo, error) {
	addApplicationServers := usecase.NewAddApplicationServers(balancerFacade.Locker,
		balancerFacade.VRRPConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)

	incomeServiceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, applicationServers)
	currentserviceInfo, err := addApplicationServers.AddNewApplicationServers(incomeServiceInfo, addApplicationServersRequestUUID)
	if err != nil {
		return incomeServiceInfo, fmt.Errorf("can't add application servers to service: %v", err)
	}
	return currentserviceInfo, nil
}

// // RemoveApplicationServersFromService ...
// func (balancerFacade *BalancerFacade) RemoveApplicationServersFromService(serviceIP, servicePort string, applicationServers map[string]string, newNWBRequestUUID string) (domain.ServiceInfo, error) {
// 	var err error
// 	var serviceInfo domain.ServiceInfo
// 	removeApplicationServers := usecase.NewRemoveApplicationServers(balancerFacade.TunnelConfig, balancerFacade.UUIDgenerator, balancerFacade.Logging)
// 	serviceInfo, err = removeApplicationServers.RemoveNewApplicationServers(serviceIP, servicePort, applicationServers, newNWBRequestUUID)
// 	if err != nil {
// 		return serviceInfo, fmt.Errorf("can't remove application servers to service: %v", err)
// 	}
// 	return serviceInfo, nil
// }

func incomeServiceDataToDomainModel(serviceIP,
	servicePort string,
	rawApplicationServers map[string]string) domain.ServiceInfo {
	applicationServers := []domain.ApplicationServer{}
	for ip, port := range rawApplicationServers {
		applicationServer := domain.ApplicationServer{
			ServerIP:   ip,
			ServerPort: port,
		}
		applicationServers = append(applicationServers, applicationServer)
	}
	return domain.ServiceInfo{
		ServiceIP:          serviceIP,
		ServicePort:        servicePort,
		ApplicationServers: applicationServers,
	}
}
