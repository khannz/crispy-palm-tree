package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const addApplicationServersName = "add-application-servers"

// AddApplicationServers ...
type AddApplicationServers struct {
	locker            *domain.Locker
	ipvsadm           *portadapter.IPVSADMEntity // so dirty
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	hc                *HeathcheckEntity
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewAddApplicationServers ...
func NewAddApplicationServers(locker *domain.Locker,
	ipvsadm *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *AddApplicationServers {
	return &AddApplicationServers{
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

// AddNewApplicationServers ...
func (addApplicationServers *AddApplicationServers) AddNewApplicationServers(newServiceInfo *domain.ServiceInfo,
	addApplicationServersUUID string) (*domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo *domain.ServiceInfo

	// graceful shutdown part start
	addApplicationServers.locker.Lock()
	defer addApplicationServers.locker.Unlock()
	addApplicationServers.gracefulShutdown.Lock()
	if addApplicationServers.gracefulShutdown.ShutdownNow {
		defer addApplicationServers.gracefulShutdown.Unlock()
		return newServiceInfo, fmt.Errorf("program got shutdown signal, job add application servers %v cancel", newServiceInfo)
	}
	addApplicationServers.gracefulShutdown.UsecasesJobs++
	addApplicationServers.gracefulShutdown.Unlock()
	defer decreaseJobs(addApplicationServers.gracefulShutdown)
	// graceful shutdown part end

	logStartUsecase(addApplicationServersName, "add new application servers to service", addApplicationServersUUID, newServiceInfo, addApplicationServers.logging)
	// FIXME: check service exist, before create tunnels
	tunnelsFilesInfo := formTunnelsFilesInfo(newServiceInfo.ApplicationServers, addApplicationServers.cacheStorage)
	logTryCreateNewTunnels(addApplicationServersName, addApplicationServersUUID, tunnelsFilesInfo, addApplicationServers.logging)
	newTunnelsFilesInfo, err := addApplicationServers.tunnelConfig.CreateTunnels(tunnelsFilesInfo, addApplicationServersUUID)
	if err != nil {
		return nil, fmt.Errorf("can't create tunnel files: %v", err)
	}
	logCreatedNewTunnels(addApplicationServersName, addApplicationServersUUID, tunnelsFilesInfo, addApplicationServers.logging)

	// need for rollback. used only service ip and port
	logTryToGetCurrentServiceInfo(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	currentServiceInfo, err := addApplicationServers.cacheStorage.GetServiceInfo(newServiceInfo, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}
	newServiceInfo.RoutingType = currentServiceInfo.RoutingType // for ipvs
	logGotCurrentServiceInfo(addApplicationServersName, addApplicationServersUUID, currentServiceInfo, addApplicationServers.logging)

	// logTryValidateForAddApplicationServers(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	// FIXME: need global rework check unique services and application servers (not at storage module!)
	allCurrentServices, err := addApplicationServers.cacheStorage.LoadAllStorageDataToDomainModel()
	if err != nil {
		return newServiceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}
	if err = checkRoutingTypeForApplicationServersValid(newServiceInfo, allCurrentServices); err != nil {
		return newServiceInfo, err
	}
	// logValidAddApplicationServers(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)

	if err := addApplicationServers.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return updatedServiceInfo, fmt.Errorf("can't update tunnel info")
	}

	logTryGenerateUpdatedServiceInfo(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	updatedServiceInfo, err = forAddApplicationServersFormUpdateServiceInfo(currentServiceInfo, newServiceInfo, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't form update service info: %v", err)
	}
	logGenerateUpdatedServiceInfo(addApplicationServersName, addApplicationServersUUID, updatedServiceInfo, addApplicationServers.logging)

	logTryIpvsadmApplicationServers(addApplicationServersName, addApplicationServersUUID, newServiceInfo.ApplicationServers, newServiceInfo.ServiceIP, newServiceInfo.ServicePort, addApplicationServers.logging)
	if err = addApplicationServers.ipvsadm.AddApplicationServersForService(newServiceInfo, addApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("Error when ipvsadm add application servers for service: %v", err)
	}
	logAddedIpvsadmApplicationServers(addApplicationServersName, addApplicationServersUUID, newServiceInfo.ApplicationServers, newServiceInfo.ServiceIP, newServiceInfo.ServicePort, addApplicationServers.logging)

	// add to cache storage
	logTryUpdateServiceInfoAtCache(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	if err = addApplicationServers.cacheStorage.UpdateServiceInfo(updatedServiceInfo, addApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}
	logUpdateServiceInfoAtCache(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)

	logTryUpdateServiceInfoAtPersistentStorage(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	if err = addApplicationServers.persistentStorage.UpdateServiceInfo(updatedServiceInfo, addApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("error when update persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)

	// TODO: why not only UpdateServiceInfo?
	if err := addApplicationServers.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return updatedServiceInfo, fmt.Errorf("can't update tunnel info")
	}

	logTryGenerateCommandsForApplicationServers(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	if err := addApplicationServers.commandGenerator.GenerateCommandsForApplicationServers(updatedServiceInfo, addApplicationServersUUID); err != nil {
		return updatedServiceInfo, fmt.Errorf("can't generate commands :%v", err)
	}
	logGeneratedCommandsForApplicationServers(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)

	logUpdateServiceAtHealtchecks(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	addApplicationServers.hc.UpdateServiceAtHealtchecks(updatedServiceInfo)
	logUpdatedServiceAtHealtchecks(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	return updatedServiceInfo, nil
}
