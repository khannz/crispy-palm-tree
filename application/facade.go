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
	GracefullShutdown *domain.GracefullShutdown
	UUIDgenerator     domain.UUIDgenerator
	Logging           *logrus.Logger
}

// NewBalancerFacade ...
func NewBalancerFacade(locker *domain.Locker,
	vrrpConfigurator domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		Locker:            locker,
		VRRPConfigurator:  vrrpConfigurator,
		CacheStorage:      cacheStorage,
		PersistentStorage: persistentStorage,
		TunnelConfig:      tunnelConfig,
		GracefullShutdown: gracefullShutdown,
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
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	serviceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, applicationServers)
	return newCreateServiceEntity.CreateService(serviceInfo, createServiceUUID)
}

// RemoveService ...
func (balancerFacade *BalancerFacade) RemoveService(serviceIP,
	servicePort string, newNWBRequestUUID string) error {
	removeService := usecase.NewRemoveServiceEntity(balancerFacade.Locker,
		balancerFacade.VRRPConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	serviceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, nil)
	return removeService.RemoveService(serviceInfo, newNWBRequestUUID)
}

// GetServices ...
func (balancerFacade *BalancerFacade) GetServices(getNWBServicesUUID string) ([]*domain.ServiceInfo, error) {
	getNWBServices := usecase.NewGetAllServices(balancerFacade.CacheStorage,
		balancerFacade.Locker,
		balancerFacade.GracefullShutdown,
		balancerFacade.Logging)
	nwbServices, err := getNWBServices.GetAllServices(getNWBServicesUUID)
	if err != nil {
		return nil, fmt.Errorf("can't get nwb services: %v", err)
	}
	return nwbServices, nil
}

// AddApplicationServers ...
func (balancerFacade *BalancerFacade) AddApplicationServers(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	addApplicationServersRequestUUID string) (*domain.ServiceInfo, error) {
	addApplicationServers := usecase.NewAddApplicationServers(balancerFacade.Locker,
		balancerFacade.VRRPConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)

	incomeServiceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, applicationServers)
	currentserviceInfo, err := addApplicationServers.AddNewApplicationServers(incomeServiceInfo, addApplicationServersRequestUUID)
	if err != nil {
		return incomeServiceInfo, fmt.Errorf("can't add application servers to service: %v", err)
	}
	return currentserviceInfo, nil
}

// RemoveApplicationServers ...
func (balancerFacade *BalancerFacade) RemoveApplicationServers(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	removeApplicationServersRequestUUID string) (*domain.ServiceInfo, error) {
	removeApplicationServers := usecase.NewRemoveApplicationServers(balancerFacade.Locker,
		balancerFacade.VRRPConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)

	incomeServiceInfo := incomeServiceDataToDomainModel(serviceIP, servicePort, applicationServers)
	currentserviceInfo, err := removeApplicationServers.RemoveApplicationServers(incomeServiceInfo, removeApplicationServersRequestUUID)
	if err != nil {
		return incomeServiceInfo, fmt.Errorf("can't remove application servers from service: %v", err)
	}
	return currentserviceInfo, nil
}

func incomeServiceDataToDomainModel(serviceIP,
	servicePort string,
	rawApplicationServers map[string]string) *domain.ServiceInfo {
	applicationServers := []domain.ApplicationServer{}
	for ip, port := range rawApplicationServers {
		applicationServer := domain.ApplicationServer{
			ServerIP:   ip,
			ServerPort: port,
		}
		applicationServers = append(applicationServers, applicationServer)
	}
	return &domain.ServiceInfo{
		ServiceIP:          serviceIP,
		ServicePort:        servicePort,
		ApplicationServers: applicationServers,
	}
}
