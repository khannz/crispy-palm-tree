package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const removeApplicationServersName = "remove-application-servers"

// RemoveApplicationServers ...
type RemoveApplicationServers struct {
	locker            *domain.Locker
	ipvsadm           domain.IPVSWorker
	cacheStorage      domain.StorageActions
	persistentStorage domain.StorageActions
	tunnelConfig      domain.TunnelMaker
	hc                *HeathcheckEntity
	gracefulShutdown  *domain.GracefulShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewRemoveApplicationServers ...
func NewRemoveApplicationServers(locker *domain.Locker,
	ipvsadm domain.IPVSWorker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveApplicationServers {
	return &RemoveApplicationServers{
		locker:            locker,
		ipvsadm:           ipvsadm,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		gracefulShutdown:  gracefulShutdown,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// RemoveApplicationServers ...
// TODO: rollbacks need refactor
func (removeApplicationServers *RemoveApplicationServers) RemoveApplicationServers(removeServiceInfo *domain.ServiceInfo,
	removeApplicationServersUUID string) (*domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo *domain.ServiceInfo

	// graceful shutdown part start
	removeApplicationServers.locker.Lock()
	defer removeApplicationServers.locker.Unlock()
	removeApplicationServers.gracefulShutdown.Lock()
	if removeApplicationServers.gracefulShutdown.ShutdownNow {
		defer removeApplicationServers.gracefulShutdown.Unlock()
		return removeServiceInfo, fmt.Errorf("program got shutdown signal, job remove application servers %v cancel", removeServiceInfo)
	}
	removeApplicationServers.gracefulShutdown.UsecasesJobs++
	removeApplicationServers.gracefulShutdown.Unlock()
	defer decreaseJobs(removeApplicationServers.gracefulShutdown)
	// graceful shutdown part end
	logStartUsecase(removeApplicationServersName, "add new application servers to service", removeApplicationServersUUID, removeServiceInfo, removeApplicationServers.logging)

	logTryPreValidateRequest(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)
	allCurrentServices, err := removeApplicationServers.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return removeServiceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}

	if isServiceExist(removeServiceInfo.ServiceIP, removeServiceInfo.ServicePort, allCurrentServices) {
		return removeServiceInfo, fmt.Errorf("service %v:%v already exist, can't create new one", removeServiceInfo.ServiceIP, removeServiceInfo.ServicePort)
	}

	logTryToGetCurrentServiceInfo(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)
	currentServiceInfo, err := removeApplicationServers.cacheStorage.GetServiceInfo(removeServiceInfo, removeApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}
	logGotCurrentServiceInfo(removeApplicationServersName, removeApplicationServersUUID, currentServiceInfo, removeApplicationServers.logging)

	if err = checkApplicationServersExistInService(removeServiceInfo.ApplicationServers, currentServiceInfo); err != nil {
		return removeServiceInfo, err
	}

	logPreValidateRequestIsOk(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)

	var tunnelsFilesInfo, oldTunnelsFilesInfo []*domain.TunnelForApplicationServer
	if currentServiceInfo.Protocol == "tcp" {
		tunnelsFilesInfo = FormTunnelsFilesInfo(removeServiceInfo.ApplicationServers, removeApplicationServers.cacheStorage)
		logTryRemoveTunnels(removeApplicationServersName, removeApplicationServersUUID, tunnelsFilesInfo, removeApplicationServers.logging)
		oldTunnelsFilesInfo, err = removeApplicationServers.tunnelConfig.RemoveTunnels(tunnelsFilesInfo, removeApplicationServersUUID)
		if err != nil {
			return nil, fmt.Errorf("can't create tunnel files: %v", err)
		}
		logRemovedTunnels(removeApplicationServersName, removeApplicationServersUUID, tunnelsFilesInfo, removeApplicationServers.logging)
	}
	logTryValidateRemoveApplicationServers(removeApplicationServersName, removeApplicationServersUUID, removeServiceInfo.ApplicationServers, removeApplicationServers.logging)
	if err = validateRemoveApplicationServers(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers); err != nil {
		return updatedServiceInfo, fmt.Errorf("validate remove application servers fail: %v", err)
	}
	logValidateRemoveApplicationServers(removeApplicationServersName, removeApplicationServersUUID, removeServiceInfo.ApplicationServers, removeApplicationServers.logging)

	updatedServiceInfo = forRemoveApplicationServersFormUpdateServiceInfo(currentServiceInfo, removeServiceInfo, removeApplicationServersUUID) // ignore check unique error
	// update for cache storage

	logTryUpdateServiceInfoAtCache(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)
	if err = removeApplicationServers.cacheStorage.UpdateServiceInfo(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}
	logUpdateServiceInfoAtCache(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)

	if currentServiceInfo.Protocol == "tcp" {
		if err = removeApplicationServers.cacheStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
			return currentServiceInfo, fmt.Errorf("can't update tunnel info in storage: %v", err)
		}
	}

	logTryRemoveIpvsadmApplicationServers(removeApplicationServersName, removeApplicationServersUUID, removeServiceInfo.ApplicationServers, removeServiceInfo.ServiceIP, removeServiceInfo.ServicePort, removeApplicationServers.logging)
	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(currentServiceInfo.ServiceIP,
		currentServiceInfo.ServicePort,
		currentServiceInfo.RoutingType,
		currentServiceInfo.BalanceType,
		currentServiceInfo.Protocol,
		removeServiceInfo.ApplicationServers)
	if err != nil {
		return currentServiceInfo, fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err = removeApplicationServers.ipvsadm.RemoveApplicationServersFromService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		removeApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("Error when ipvsadm remove application servers from service: %v", err)
	}
	logRemovedIpvsadmApplicationServers(removeApplicationServersName, removeApplicationServersUUID, removeServiceInfo.ApplicationServers, removeServiceInfo.ServiceIP, removeServiceInfo.ServicePort, removeApplicationServers.logging)

	logTryUpdateServiceInfoAtPersistentStorage(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)
	if err = removeApplicationServers.persistentStorage.UpdateServiceInfo(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("Error when update persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)

	if currentServiceInfo.Protocol == "tcp" {
		if err = removeApplicationServers.persistentStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
			return currentServiceInfo, fmt.Errorf("can't update tunnel info in storage: %v", err)
		}
	}

	logUpdateServiceAtHealtchecks(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)
	if err = removeApplicationServers.hc.UpdateServiceAtHealtchecks(updatedServiceInfo); err != nil {
		return updatedServiceInfo, fmt.Errorf("application server removed, butan error occurred when removing it from the healtchecks: %v", err)
	}
	logUpdatedServiceAtHealtchecks(removeApplicationServersName, removeApplicationServersUUID, removeApplicationServers.logging)

	return updatedServiceInfo, nil
}
