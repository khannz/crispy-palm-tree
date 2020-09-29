package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const createServiceName = "create-service"

// CreateServiceEntity ...
type CreateServiceEntity struct {
	locker            *domain.Locker
	ipvsadm           domain.IPVSWorker
	cacheStorage      domain.StorageActions
	persistentStorage domain.StorageActions
	tunnelConfig      domain.TunnelMaker
	hc                domain.HeathcheckWorker
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
	logging           *logrus.Logger
}

// NewCreateServiceEntity ...
func NewCreateServiceEntity(locker *domain.Locker,
	ipvsadm domain.IPVSWorker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc domain.HeathcheckWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *CreateServiceEntity {
	return &CreateServiceEntity{
		locker:            locker,
		ipvsadm:           ipvsadm,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		commandGenerator:  commandGenerator,
		gracefulShutdown:  gracefulShutdown,
		logging:           logging,
	}
}

// CreateService ...
func (createService *CreateServiceEntity) CreateService(serviceInfo *domain.ServiceInfo,
	createServiceID string) (*domain.ServiceInfo, error) {
	// graceful shutdown part start
	createService.locker.Lock()
	defer createService.locker.Unlock()
	createService.gracefulShutdown.Lock()
	if createService.gracefulShutdown.ShutdownNow {
		defer createService.gracefulShutdown.Unlock()
		return serviceInfo, fmt.Errorf("program got shutdown signal, job create service %v cancel", serviceInfo)
	}
	createService.gracefulShutdown.UsecasesJobs++
	createService.gracefulShutdown.Unlock()
	defer decreaseJobs(createService.gracefulShutdown)
	// graceful shutdown part end
	logStartUsecase(createServiceName, "add new application servers to service", createServiceID, serviceInfo, createService.logging)

	logTryPreValidateRequest(createServiceName, createServiceID, createService.logging)
	allCurrentServices, err := createService.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return serviceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}

	if isServiceExist(serviceInfo.ServiceIP, serviceInfo.ServicePort, allCurrentServices) {
		return serviceInfo, fmt.Errorf("service %v:%v already exist, can't create new one", serviceInfo.ServiceIP, serviceInfo.ServicePort)
	}

	if err = checkServiceIPAndPortUnique(serviceInfo.ServiceIP, serviceInfo.ServicePort, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	if err = checkApplicationServersIPAndPortUnique(serviceInfo.ApplicationServers, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	if err = checkRoutingTypeForApplicationServersValid(serviceInfo, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	logPreValidateRequestIsOk(createServiceName, createServiceID, createService.logging)

	var tunnelsFilesInfo, newTunnelsFilesInfo []*domain.TunnelForApplicationServer
	if serviceInfo.Protocol == "tcp" { // TODO: too many if's, that dirty
		tunnelsFilesInfo = FormTunnelsFilesInfo(serviceInfo.ApplicationServers, createService.cacheStorage)
		logTryCreateNewTunnels(createServiceName, createServiceID, tunnelsFilesInfo, createService.logging)
		newTunnelsFilesInfo, err = createService.tunnelConfig.CreateTunnels(tunnelsFilesInfo, createServiceID)
		if err != nil {
			return serviceInfo, fmt.Errorf("can't create tunnel files: %v", err)
		}
		logCreatedNewTunnels(createServiceName, createServiceID, tunnelsFilesInfo, createService.logging)
	}

	logTryCreateIPVSService(createServiceName, createServiceID, serviceInfo.ApplicationServers, serviceInfo.ServiceIP, serviceInfo.ServicePort, createService.logging)
	vip, port, routingType, balanceType, protocol, err := domain.PrepareServiceForIPVS(serviceInfo.ServiceIP,
		serviceInfo.ServicePort,
		serviceInfo.RoutingType,
		serviceInfo.BalanceType,
		serviceInfo.Protocol)
	if err != nil {
		return serviceInfo, fmt.Errorf("error prepare data for IPVS: %v", err)
	}
	if err := createService.ipvsadm.CreateService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		nil,
		createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("Error when ipvsadm create service: %v", err)
	}
	logCreatedIPVSService(createServiceName, createServiceID, serviceInfo.ApplicationServers, serviceInfo.ServiceIP, serviceInfo.ServicePort, createService.logging)

	// add to cache storage
	logTryUpdateServiceInfoAtCache(createServiceName, createServiceID, createService.logging)
	if err := createService.cacheStorage.NewServiceInfoToStorage(serviceInfo, createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}
	logUpdateServiceInfoAtCache(createServiceName, createServiceID, createService.logging)

	if serviceInfo.Protocol == "tcp" {
		if err := createService.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
		}
	}

	logTryUpdateServiceInfoAtPersistentStorage(createServiceName, createServiceID, createService.logging)
	if err = createService.persistentStorage.NewServiceInfoToStorage(serviceInfo, createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("Error when save to persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(createServiceName, createServiceID, createService.logging)

	if serviceInfo.Protocol == "tcp" {
		if err := createService.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return serviceInfo, fmt.Errorf("can't add to persistent storage :%v", err)
		}
	}
	logTryGenerateCommandsForApplicationServers(createServiceName, createServiceID, createService.logging)
	if err := createService.commandGenerator.GenerateCommandsForApplicationServers(serviceInfo, createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("can't generate commands :%v", err)
	}
	logGeneratedCommandsForApplicationServers(createServiceName, createServiceID, createService.logging)

	logUpdateServiceAtHealtchecks(createServiceName, createServiceID, createService.logging)
	createService.hc.NewServiceToHealtchecks(serviceInfo)
	logUpdatedServiceAtHealtchecks(createServiceName, createServiceID, createService.logging)
	return serviceInfo, nil
}
