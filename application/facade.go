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
	CommandGenerator    domain.CommandGenerator
	GracefulShutdown    *domain.GracefulShutdown
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
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		Locker:              locker,
		IPVSADMConfigurator: ipvsadmConfigurator,
		CacheStorage:        cacheStorage,
		PersistentStorage:   persistentStorage,
		TunnelConfig:        tunnelConfig,
		HeathcheckEntity:    hc,
		CommandGenerator:    commandGenerator,
		GracefulShutdown:    gracefulShutdown,
		UUIDgenerator:       uuidGenerator,
		Logging:             logging,
	}
}

// CreateService ...
func (balancerFacade *BalancerFacade) CreateService(createService *NewServiceInfo,
	createServiceUUID string) (*domain.ServiceInfo, error) {
	newCreateServiceEntity := usecase.NewCreateServiceEntity(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.CommandGenerator,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)
	appSvrs := []*domain.ApplicationServer{}
	for _, appSrvr := range createService.ApplicationServers {
		arrayOfAdvancedHealthcheckParameters := []domain.AdvancedHealthcheckParameters{}
		for _, aHP := range appSrvr.ServerHealthcheck.AdvancedHealthcheckParameters {
			advancedHealthcheckParameter := domain.AdvancedHealthcheckParameters{
				NearFieldsMode:  aHP.NearFieldsMode,
				UserDefinedData: aHP.UserDefinedData,
			}
			arrayOfAdvancedHealthcheckParameters = append(arrayOfAdvancedHealthcheckParameters, advancedHealthcheckParameter)
		}

		hcA := domain.ServerHealthcheck{
			TypeOfCheck:                   appSrvr.ServerHealthcheck.TypeOfCheck,
			HealthcheckAddress:            appSrvr.ServerHealthcheck.HealthcheckAddress,
			AdvancedHealthcheckParameters: arrayOfAdvancedHealthcheckParameters,
		}
		as := &domain.ApplicationServer{
			ServerIP:          appSrvr.ServerIP,
			ServerPort:        appSrvr.ServerPort,
			ServerHealthcheck: hcA,
			IsUp:              false,
		}
		appSvrs = append(appSvrs, as)
	}
	hcS := domain.ServiceHealthcheck{
		Type:                 createService.Healtcheck.Type,
		Timeout:              createService.Healtcheck.Timeout,
		RepeatHealthcheck:    createService.Healtcheck.RepeatHealthcheck,
		PercentOfAlivedForUp: createService.Healtcheck.PercentOfAlivedForUp,
	}

	serviceInfo := &domain.ServiceInfo{
		ServiceIP:          createService.ServiceIP,
		ServicePort:        createService.ServicePort,
		ApplicationServers: appSvrs,
		Healthcheck:        hcS,
		BalanceType:        createService.BalanceType,
		RoutingType:        createService.RoutingType,
		IsUp:               false,
	}
	return newCreateServiceEntity.CreateService(serviceInfo, createServiceUUID)
}

// RemoveService ...
func (balancerFacade *BalancerFacade) RemoveService(removeServiceRequest *RemoveServiceInfo,
	newNWBRequestUUID string) error {
	removeService := usecase.NewRemoveServiceEntity(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.GracefulShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.HeathcheckEntity,
		balancerFacade.Logging)
	serviceInfo := &domain.ServiceInfo{ServiceIP: removeServiceRequest.ServiceIP, ServicePort: removeServiceRequest.ServicePort}
	return removeService.RemoveService(serviceInfo, newNWBRequestUUID)
}

// GetServices ...
func (balancerFacade *BalancerFacade) GetServices(getNWBServicesUUID string) ([]*domain.ServiceInfo, error) {
	getNWBServices := usecase.NewGetAllServices(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.GracefulShutdown,
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
		balancerFacade.CommandGenerator,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)

	appSvrs := []*domain.ApplicationServer{}
	for _, appSrvr := range addApplicationServersRequest.ApplicationServers {
		arrayOfAdvancedHealthcheckParameters := []domain.AdvancedHealthcheckParameters{}
		for _, aHP := range appSrvr.ServerHealthcheck.AdvancedHealthcheckParameters {
			advancedHealthcheckParameter := domain.AdvancedHealthcheckParameters{
				NearFieldsMode:  aHP.NearFieldsMode,
				UserDefinedData: aHP.UserDefinedData,
			}
			arrayOfAdvancedHealthcheckParameters = append(arrayOfAdvancedHealthcheckParameters, advancedHealthcheckParameter)
		}

		hcA := domain.ServerHealthcheck{
			TypeOfCheck:                   appSrvr.ServerHealthcheck.TypeOfCheck,
			HealthcheckAddress:            appSrvr.ServerHealthcheck.HealthcheckAddress,
			AdvancedHealthcheckParameters: arrayOfAdvancedHealthcheckParameters,
		}
		as := &domain.ApplicationServer{
			ServerIP:          appSrvr.ServerIP,
			ServerPort:        appSrvr.ServerPort,
			ServerHealthcheck: hcA,
			IsUp:              false,
		}
		appSvrs = append(appSvrs, as)
	}

	incomeServiceInfo := &domain.ServiceInfo{
		ServiceIP:          addApplicationServersRequest.ServiceIP,
		ServicePort:        addApplicationServersRequest.ServicePort,
		ApplicationServers: appSvrs,
	}
	currentserviceInfo, err := addApplicationServers.AddNewApplicationServers(incomeServiceInfo, addApplicationServersRequestUUID)
	if err != nil {
		return incomeServiceInfo, fmt.Errorf("can't add application servers to service: %v", err)
	}
	return currentserviceInfo, nil
}

// RemoveApplicationServers ...
func (balancerFacade *BalancerFacade) RemoveApplicationServers(removeApplicationServersRequest *RemoveApplicationServersRequest,
	removeApplicationServersRequestUUID string) (*domain.ServiceInfo, error) {
	removeApplicationServers := usecase.NewRemoveApplicationServers(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.GracefulShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	appSvrs := []*domain.ApplicationServer{}
	for _, appSrvr := range removeApplicationServersRequest.ApplicationServers {
		arrayOfAdvancedHealthcheckParameters := []domain.AdvancedHealthcheckParameters{}
		for _, aHP := range appSrvr.ServerHealthcheck.AdvancedHealthcheckParameters {
			advancedHealthcheckParameter := domain.AdvancedHealthcheckParameters{
				NearFieldsMode:  aHP.NearFieldsMode,
				UserDefinedData: aHP.UserDefinedData,
			}
			arrayOfAdvancedHealthcheckParameters = append(arrayOfAdvancedHealthcheckParameters, advancedHealthcheckParameter)
		}

		hcA := domain.ServerHealthcheck{
			TypeOfCheck:                   appSrvr.ServerHealthcheck.TypeOfCheck,
			HealthcheckAddress:            appSrvr.ServerHealthcheck.HealthcheckAddress,
			AdvancedHealthcheckParameters: arrayOfAdvancedHealthcheckParameters,
		}
		as := &domain.ApplicationServer{
			ServerIP:          appSrvr.ServerIP,
			ServerPort:        appSrvr.ServerPort,
			ServerHealthcheck: hcA,
		}
		appSvrs = append(appSvrs, as)
	}

	incomeServiceInfo := &domain.ServiceInfo{
		ServiceIP:          removeApplicationServersRequest.ServiceIP,
		ServicePort:        removeApplicationServersRequest.ServicePort,
		ApplicationServers: appSvrs,
	}

	currentserviceInfo, err := removeApplicationServers.RemoveApplicationServers(incomeServiceInfo, removeApplicationServersRequestUUID)
	if err != nil {
		return incomeServiceInfo, fmt.Errorf("can't remove application servers from service: %v", err)
	}
	return currentserviceInfo, nil
}

// GetServiceState ...
func (balancerFacade *BalancerFacade) GetServiceState(getServiceStateRequest *GetServiceStateRequest) (*domain.ServiceInfo, error) {
	getServiceStateEntity := usecase.NewGetServiceStateEntity(balancerFacade.Locker,
		balancerFacade.CacheStorage,
		balancerFacade.GracefulShutdown,
		balancerFacade.Logging)
	incomeServiceInfo := &domain.ServiceInfo{
		ServiceIP:   getServiceStateRequest.ServiceIP,
		ServicePort: getServiceStateRequest.ServicePort,
	}
	return getServiceStateEntity.GetServiceState(incomeServiceInfo, getServiceStateRequest.ID)
}

// ModifyService ...
func (balancerFacade *BalancerFacade) ModifyService(modifyService *ModifyServiceInfo,
	modifyServiceUUID string) (*domain.ServiceInfo, error) {
	newModifyServiceEntity := usecase.NewModifyServiceEntity(balancerFacade.Locker,
		balancerFacade.IPVSADMConfigurator,
		balancerFacade.CacheStorage,
		balancerFacade.PersistentStorage,
		balancerFacade.TunnelConfig,
		balancerFacade.HeathcheckEntity,
		balancerFacade.CommandGenerator,
		balancerFacade.GracefulShutdown,
		balancerFacade.UUIDgenerator,
		balancerFacade.Logging)
	appSvrs := []*domain.ApplicationServer{}
	for _, appSrvr := range modifyService.ApplicationServers {
		hcA := domain.ServerHealthcheck{HealthcheckAddress: appSrvr.ServerHealthcheck.HealthcheckAddress}
		as := &domain.ApplicationServer{
			ServerIP:          appSrvr.ServerIP,
			ServerPort:        appSrvr.ServerPort,
			ServerHealthcheck: hcA,
			IsUp:              false,
		}
		appSvrs = append(appSvrs, as)
	}
	hcS := domain.ServiceHealthcheck{
		Type:                 modifyService.Healtcheck.Type,
		Timeout:              modifyService.Healtcheck.Timeout,
		RepeatHealthcheck:    modifyService.Healtcheck.RepeatHealthcheck,
		PercentOfAlivedForUp: modifyService.Healtcheck.PercentOfAlivedForUp,
	}

	serviceInfo := &domain.ServiceInfo{
		ServiceIP:          modifyService.ServiceIP,
		ServicePort:        modifyService.ServicePort,
		ApplicationServers: appSvrs,
		Healthcheck:        hcS,
		BalanceType:        modifyService.BalanceType,
		RoutingType:        modifyService.RoutingType,
		IsUp:               false,
	}
	return newModifyServiceEntity.ModifyService(serviceInfo, modifyServiceUUID)
}
