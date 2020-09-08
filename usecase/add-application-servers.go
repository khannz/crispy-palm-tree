package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const addApplicationServersName = "add-application-servers"

// AddApplicationServers ...
type AddApplicationServers struct {
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

// NewAddApplicationServers ...
func NewAddApplicationServers(locker *domain.Locker,
	ipvsadm domain.IPVSWorker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc domain.HeathcheckWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
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
	logTryPreValidateRequest(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	allCurrentServices, err := addApplicationServers.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return newServiceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}
	if !isServiceExist(newServiceInfo.ServiceIP, newServiceInfo.ServicePort, allCurrentServices) {
		return newServiceInfo, fmt.Errorf("service %v:%v does not exist, can't add application servers", newServiceInfo.ServiceIP, newServiceInfo.ServicePort)
	}

	if err = checkApplicationServersIPAndPortUnique(newServiceInfo.ApplicationServers, allCurrentServices); err != nil {
		return newServiceInfo, err
	}

	logTryToGetCurrentServiceInfo(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	currentServiceInfo, err := addApplicationServers.cacheStorage.GetServiceInfo(newServiceInfo, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}
	newServiceInfo.RoutingType = currentServiceInfo.RoutingType // for ipvs and for check routing type is valid
	logGotCurrentServiceInfo(addApplicationServersName, addApplicationServersUUID, currentServiceInfo, addApplicationServers.logging)

	if err = checkRoutingTypeForApplicationServersValid(newServiceInfo, allCurrentServices); err != nil {
		return newServiceInfo, err
	}
	logPreValidateRequestIsOk(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)

	tunnelsFilesInfo := FormTunnelsFilesInfo(newServiceInfo.ApplicationServers, addApplicationServers.cacheStorage)
	logTryCreateNewTunnels(addApplicationServersName, addApplicationServersUUID, tunnelsFilesInfo, addApplicationServers.logging)
	newTunnelsFilesInfo, err := addApplicationServers.tunnelConfig.CreateTunnels(tunnelsFilesInfo, addApplicationServersUUID)
	if err != nil {
		return nil, fmt.Errorf("can't create tunnel files: %v", err)
	}
	logCreatedNewTunnels(addApplicationServersName, addApplicationServersUUID, tunnelsFilesInfo, addApplicationServers.logging)

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

	if err := addApplicationServers.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return updatedServiceInfo, fmt.Errorf("can't update tunnel info")
	}

	logTryGenerateCommandsForApplicationServers(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	if err := addApplicationServers.commandGenerator.GenerateCommandsForApplicationServers(updatedServiceInfo, addApplicationServersUUID); err != nil {
		return updatedServiceInfo, fmt.Errorf("can't generate commands :%v", err)
	}
	logGeneratedCommandsForApplicationServers(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)

	logUpdateServiceAtHealtchecks(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	if err = addApplicationServers.hc.UpdateServiceAtHealtchecks(updatedServiceInfo); err != nil {
		return updatedServiceInfo, fmt.Errorf("application server added, but not activated, an error occurred when adding to the healtchecks: %v", err)
	}
	logUpdatedServiceAtHealtchecks(addApplicationServersName, addApplicationServersUUID, addApplicationServers.logging)
	return updatedServiceInfo, nil
}
