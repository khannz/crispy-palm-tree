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
	hc                *HeathcheckEntity
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewCreateServiceEntity ...
func NewCreateServiceEntity(locker *domain.Locker,
	ipvsadm domain.IPVSWorker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	uuidGenerator domain.UUIDgenerator,
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
		uuidGenerator:     uuidGenerator,
	}
}

// CreateService ...
func (createService *CreateServiceEntity) CreateService(serviceInfo *domain.ServiceInfo,
	createServiceUUID string) (*domain.ServiceInfo, error) {
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
	logStartUsecase(createServiceName, "add new application servers to service", createServiceUUID, serviceInfo, createService.logging)

	logTryPreValidateRequest(createServiceName, createServiceUUID, createService.logging)
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

	logPreValidateRequestIsOk(createServiceName, createServiceUUID, createService.logging)

	tunnelsFilesInfo := FormTunnelsFilesInfo(serviceInfo.ApplicationServers, createService.cacheStorage)
	logTryCreateNewTunnels(createServiceName, createServiceUUID, tunnelsFilesInfo, createService.logging)
	newTunnelsFilesInfo, err := createService.tunnelConfig.CreateTunnels(tunnelsFilesInfo, createServiceUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't create tunnel files: %v", err)
	}
	logCreatedNewTunnels(createServiceName, createServiceUUID, tunnelsFilesInfo, createService.logging)
	// add to cache storage
	logTryUpdateServiceInfoAtCache(createServiceName, createServiceUUID, createService.logging)
	if err := createService.cacheStorage.NewServiceInfoToStorage(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}
	logUpdateServiceInfoAtCache(createServiceName, createServiceUUID, createService.logging)

	// TODO: why not in NewServiceInfoToStorage? double?
	if err := createService.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}

	logTryCreateIPVSService(createServiceName, createServiceUUID, serviceInfo.ApplicationServers, serviceInfo.ServiceIP, serviceInfo.ServicePort, createService.logging)
	if err := createService.ipvsadm.CreateService(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("Error when ipvsadm create service: %v", err)
	}
	logCreatedIPVSService(createServiceName, createServiceUUID, serviceInfo.ApplicationServers, serviceInfo.ServiceIP, serviceInfo.ServicePort, createService.logging)

	logTryUpdateServiceInfoAtPersistentStorage(createServiceName, createServiceUUID, createService.logging)
	if err = createService.persistentStorage.NewServiceInfoToStorage(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("Error when save to persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(createServiceName, createServiceUUID, createService.logging)

	// TODO: double?
	if err := createService.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}

	logTryGenerateCommandsForApplicationServers(createServiceName, createServiceUUID, createService.logging)
	if err := createService.commandGenerator.GenerateCommandsForApplicationServers(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("can't generate commands :%v", err)
	}
	logGeneratedCommandsForApplicationServers(createServiceName, createServiceUUID, createService.logging)

	logUpdateServiceAtHealtchecks(createServiceName, createServiceUUID, createService.logging)
	createService.hc.NewServiceToHealtchecks(serviceInfo)
	logUpdatedServiceAtHealtchecks(createServiceName, createServiceUUID, createService.logging)
	return serviceInfo, nil
}
