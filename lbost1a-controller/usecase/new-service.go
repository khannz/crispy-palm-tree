package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const newServiceName = "new service"

// NewServiceEntity ...
type NewServiceEntity struct {
	locker            *domain.Locker
	cacheStorage      domain.StorageActions
	persistentStorage domain.StorageActions
	tunnelConfig      domain.TunnelMaker
	hc                domain.HCWorker
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
	logging           *logrus.Logger
}

// NewNewServiceEntity ... // TODO: naming
func NewNewServiceEntity(locker *domain.Locker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc domain.HCWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *NewServiceEntity {
	return &NewServiceEntity{
		locker:            locker,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		commandGenerator:  commandGenerator,
		gracefulShutdown:  gracefulShutdown,
		logging:           logging,
	}
}

// NewService ...
func (createService *NewServiceEntity) NewService(serviceInfo *domain.ServiceInfo,
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
	logStartUsecase(newServiceName, "add new application servers to service", createServiceID, serviceInfo, createService.logging)

	logTryPreValidateRequest(newServiceName, createServiceID, createService.logging)
	allCurrentServices, err := createService.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return serviceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}

	if isServiceExist(serviceInfo.IP, serviceInfo.Port, allCurrentServices) {
		return serviceInfo, fmt.Errorf("service %v:%v already exist, can't create new one", serviceInfo.IP, serviceInfo.Port)
	}

	if err = checkIPAndPortUnique(serviceInfo.IP, serviceInfo.Port, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	if err = checkApplicationServersIPAndPortUnique(serviceInfo.ApplicationServers, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	if err = checkRoutingTypeForApplicationServersValid(serviceInfo, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	logPreValidateRequestIsOk(newServiceName, createServiceID, createService.logging)

	var tunnelsFilesInfo, newTunnelsFilesInfo []*domain.TunnelForApplicationServer
	if serviceInfo.Protocol == "tcp" { // TODO: too many if's, that dirty
		tunnelsFilesInfo = FormTunnelsFilesInfo(serviceInfo.ApplicationServers, createService.cacheStorage)
		logTryCreateNewTunnels(newServiceName, createServiceID, tunnelsFilesInfo, createService.logging)
		newTunnelsFilesInfo, err = createService.tunnelConfig.CreateTunnels(tunnelsFilesInfo, createServiceID)
		if err != nil {
			return serviceInfo, fmt.Errorf("can't create tunnel files: %v", err)
		}
		logCreatedNewTunnels(newServiceName, createServiceID, tunnelsFilesInfo, createService.logging)
	}

	// add to cache storage
	logTryUpdateServiceInfoAtCache(newServiceName, createServiceID, createService.logging)
	if err := createService.cacheStorage.NewServiceInfoToStorage(serviceInfo, createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}
	logUpdateServiceInfoAtCache(newServiceName, createServiceID, createService.logging)

	if serviceInfo.Protocol == "tcp" {
		if err := createService.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
		}
	}

	logTryUpdateServiceInfoAtPersistentStorage(newServiceName, createServiceID, createService.logging)
	if err = createService.persistentStorage.NewServiceInfoToStorage(serviceInfo, createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("Error when save to persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(newServiceName, createServiceID, createService.logging)

	if serviceInfo.Protocol == "tcp" {
		if err := createService.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return serviceInfo, fmt.Errorf("can't add to persistent storage :%v", err)
		}
	}
	logTryGenerateCommandsForApplicationServers(newServiceName, createServiceID, createService.logging)
	if err := createService.commandGenerator.GenerateCommandsForApplicationServers(serviceInfo, createServiceID); err != nil {
		return serviceInfo, fmt.Errorf("can't generate commands :%v", err)
	}
	logGeneratedCommandsForApplicationServers(newServiceName, createServiceID, createService.logging)

	logUpdateServiceAtHealtchecks(newServiceName, createServiceID, createService.logging)
	if err = createService.hc.NewServiceToHealtchecks(serviceInfo); err != nil {
		return serviceInfo, fmt.Errorf("error when change service in healthcheck: %v", err)
	}
	logUpdatedServiceAtHealtchecks(newServiceName, createServiceID, createService.logging)
	return serviceInfo, nil
}
