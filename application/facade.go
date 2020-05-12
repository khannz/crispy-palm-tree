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
	Locker              *domain.Locker
	IPVSADMConfigurator *portadapter.IPVSADMEntity
	CacheStorage        *portadapter.StorageEntity // so dirty
	PersistentStorage   *portadapter.StorageEntity // so dirty
	TunnelConfig        domain.TunnelMaker
	HeathcheckEntity    *usecase.HeathcheckEntity
	GracefullShutdown   *domain.GracefullShutdown
	UUIDgenerator       domain.UUIDgenerator
	Logging             *logrus.Logger
}

// NewBalancerFacade ...
func NewBalancerFacade(locker *domain.Locker,
	ipvsadmConfigurator *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	hc *usecase.HeathcheckEntity,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		Locker:              locker,
		IPVSADMConfigurator: ipvsadmConfigurator,
		CacheStorage:        cacheStorage,
		PersistentStorage:   persistentStorage,
		TunnelConfig:        tunnelConfig,
		HeathcheckEntity:    hc,
		GracefullShutdown:   gracefullShutdown,
		UUIDgenerator:       uuidGenerator,
		Logging:             logging,
	}
}

// CreateService ...
func (balancerFacade *BalancerFacade) CreateService(createService *NewBalanceInfo,
	createServiceUUID string) error {
	newCreateServiceEntity := usecase.NewCreateServiceEntity(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	appSvrs := []*domain.ApplicationServer{}
	for _, appSrvr := range createService.ApplicationServers {
		hcA := domain.ServerHealthcheck{HealthcheckAddress: appSrvr.ServerHealthcheck.HealthcheckAddress}
		as := &domain.ApplicationServer{
			ServerIP:          appSrvr.ServerIP,
			ServerPort:        appSrvr.ServerPort,
			ServerHealthcheck: hcA,
		}
		appSvrs = append(appSvrs, as)
	}
	hcS := domain.ServiceHealthcheck{
		Type:    createService.Healtcheck.Type,
		Timeout: createService.Healtcheck.Timeout,
	}

	serviceInfo := &domain.ServiceInfo{
		ServiceIP:          createService.ServiceIP,
		ServicePort:        createService.ServicePort,
		ApplicationServers: appSvrs,
		Healthcheck:        hcS,
	}
	return newCreateServiceEntity.CreateService(serviceInfo, createServiceUUID)
}

// RemoveService ...
func (balancerFacade *BalancerFacade) RemoveService(serviceIP,
	servicePort string, newNWBRequestUUID string) error {
	removeService := usecase.NewRemoveServiceEntity(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.HeathcheckEntity,
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
func (balancerFacade *BalancerFacade) AddApplicationServers(addApplicationServersRequest *AddApplicationServersRequest,
	addApplicationServersRequestUUID string) (*domain.ServiceInfo, error) {
	addApplicationServers := usecase.NewAddApplicationServers(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefullShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)

	appSvrs := []*domain.ApplicationServer{}
	for _, appSrvr := range addApplicationServersRequest.ApplicationServers {
		hcA := domain.ServerHealthcheck{HealthcheckAddress: appSrvr.ServerHealthcheck.HealthcheckAddress}
		as := &domain.ApplicationServer{
			ServerIP:          appSrvr.ServerIP,
			ServerPort:        appSrvr.ServerPort,
			ServerHealthcheck: hcA,
		}
		appSvrs = append(appSvrs, as)
	}

	hcS := domain.ServiceHealthcheck{
		Type:    addApplicationServersRequest.Healtcheck.Type,
		Timeout: addApplicationServersRequest.Healtcheck.Timeout,
	}
	incomeServiceInfo := &domain.ServiceInfo{
		ServiceIP:          addApplicationServersRequest.ServiceIP,
		ServicePort:        addApplicationServersRequest.ServicePort,
		ApplicationServers: appSvrs,
		Healthcheck:        hcS,
	}
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
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
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
	applicationServers := []*domain.ApplicationServer{}
	for ip, port := range rawApplicationServers {
		applicationServer := &domain.ApplicationServer{
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
