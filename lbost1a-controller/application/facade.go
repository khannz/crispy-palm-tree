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
	CacheStorage      *portadapter.StorageEntity // so dirty
	PersistentStorage *portadapter.StorageEntity // so dirty
	TunnelConfig      domain.TunnelMaker
	HeathcheckEntity  domain.HCWorker
	CommandGenerator  domain.CommandGenerator
	GracefulShutdown  *domain.GracefulShutdown
	IDgenerator       domain.IDgenerator
	Logging           *logrus.Logger
}

// NewBalancerFacade ...
func NewBalancerFacade(locker *domain.Locker,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	hc domain.HCWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		Locker:            locker,
		CacheStorage:      cacheStorage,
		PersistentStorage: persistentStorage,
		TunnelConfig:      tunnelConfig,
		HeathcheckEntity:  hc,
		CommandGenerator:  commandGenerator,
		GracefulShutdown:  gracefulShutdown,
		IDgenerator:       idGenerator,
		Logging:           logging,
	}
}

// NewService ...
func (balancerFacade *BalancerFacade) NewService(newService *Service,
	newServiceID string) (*Service, error) {
	newNewServiceEntity := usecase.NewNewServiceEntity(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.CommandGenerator,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)
	serviceInfo := convertRestServiceToDomainServiceInfo(newService)

	resultNewServiceInfo, err := newNewServiceEntity.NewService(serviceInfo, newServiceID)
	if err != nil {
		return newService, err
	}
	return convertDomainServiceInfoToRestService(resultNewServiceInfo), nil
}

// RemoveService ...
func (balancerFacade *BalancerFacade) RemoveService(ip, port, newNWBRequestID string) error {
	removeService := usecase.NewRemoveServiceEntity(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)
	serviceInfo := &domain.ServiceInfo{Address: ip + ":" + port, IP: ip, Port: port}
	return removeService.RemoveService(serviceInfo, newNWBRequestID)
}

// GetServices ...
func (balancerFacade *BalancerFacade) GetServices(getNWBServicesID string) ([]*domain.ServiceInfo, error) {
	getNWBServices := usecase.NewGetAllServices(balancerFacade.Locker,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)
	nwbServices, err := getNWBServices.GetAllServices(getNWBServicesID)
	if err != nil {
		return nil, fmt.Errorf("can't get nwb services: %v", err)
	}
	return nwbServices, nil
}

// AddApplicationServers ...
func (balancerFacade *BalancerFacade) AddApplicationServers(addApplicationServersRequest *Service,
	addApplicationServersRequestID string) (*Service, error) {
	addApplicationServers := usecase.NewAddApplicationServers(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.CommandGenerator,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)

	incomeServiceInfo := convertRestServiceToDomainServiceInfo(addApplicationServersRequest)
	currentserviceInfo, err := addApplicationServers.AddNewApplicationServers(incomeServiceInfo, addApplicationServersRequestID)
	if err != nil {
		return addApplicationServersRequest, fmt.Errorf("can't add application servers to service: %v", err)
	}
	return convertDomainServiceInfoToRestService(currentserviceInfo), nil
}

// RemoveApplicationServers ...
func (balancerFacade *BalancerFacade) RemoveApplicationServers(removeApplicationServersRequest *Service,
	removeApplicationServersRequestID string) (*Service, error) {
	removeApplicationServers := usecase.NewRemoveApplicationServers(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)

	incomeServiceInfo := convertRestServiceToDomainServiceInfo(removeApplicationServersRequest)

	resultCurrentServiceInfo, err := removeApplicationServers.RemoveApplicationServers(incomeServiceInfo, removeApplicationServersRequestID)
	if err != nil {
		return removeApplicationServersRequest, fmt.Errorf("can't remove application servers from service: %v", err)
	}
	return convertDomainServiceInfoToRestService(resultCurrentServiceInfo), nil
}

// GetServiceState ...
func (balancerFacade *BalancerFacade) GetServiceState(ip, port, getServiceRequestID string) (*Service, error) {
	getServiceStateEntity := usecase.NewGetServiceStateEntity(balancerFacade.Locker,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)
	incomeServiceInfo := &domain.ServiceInfo{
		Address: ip + ":" + port,
		IP:      ip,
		Port:    port,
	}
	serviceInfo, err := getServiceStateEntity.GetServiceState(incomeServiceInfo, getServiceRequestID)
	if err != nil {
		return convertDomainServiceInfoToRestService(incomeServiceInfo), err
	}
	return convertDomainServiceInfoToRestService(serviceInfo), nil
}

// ModifyService ...
func (balancerFacade *BalancerFacade) ModifyService(modifyService *Service,
	modifyServiceID string) (*Service, error) {
	newModifyServiceEntity := usecase.NewModifyServiceEntity(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.CommandGenerator,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)

	serviceInfo := convertRestServiceToDomainServiceInfo(modifyService)
	updatedServiceInfo, err := newModifyServiceEntity.ModifyService(serviceInfo, modifyServiceID)
	if err != nil {
		return modifyService, err
	}
	return convertDomainServiceInfoToRestService(updatedServiceInfo), nil
}
