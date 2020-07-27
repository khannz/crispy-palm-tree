package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const removeServiceName = "remove-service"

// RemoveServiceEntity ...
type RemoveServiceEntity struct {
	locker            *domain.Locker
	ipvsadm           *portadapter.IPVSADMEntity
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	gracefulShutdown  *domain.GracefulShutdown
	uuidGenerator     domain.UUIDgenerator
	hc                *HeathcheckEntity
	logging           *logrus.Logger
}

// NewRemoveServiceEntity ...
func NewRemoveServiceEntity(locker *domain.Locker,
	ipvsadm *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
	gracefulShutdown *domain.GracefulShutdown,
	uuidGenerator domain.UUIDgenerator,
	hc *HeathcheckEntity,
	logging *logrus.Logger) *RemoveServiceEntity {
	return &RemoveServiceEntity{
		locker:            locker,
		ipvsadm:           ipvsadm,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		gracefulShutdown:  gracefulShutdown,
		uuidGenerator:     uuidGenerator,
		hc:                hc,
		logging:           logging,
	}
}

// RemoveService ...
// FIXME: rollbacks need refactor
func (removeServiceEntity *RemoveServiceEntity) RemoveService(serviceInfo *domain.ServiceInfo,
	removeServiceUUID string) error {
	var err error

	// graceful shutdown part start
	removeServiceEntity.locker.Lock()
	defer removeServiceEntity.locker.Unlock()
	removeServiceEntity.gracefulShutdown.Lock()
	if removeServiceEntity.gracefulShutdown.ShutdownNow {
		defer removeServiceEntity.gracefulShutdown.Unlock()
		return fmt.Errorf("program got shutdown signal, job remove service %v cancel", serviceInfo)
	}
	removeServiceEntity.gracefulShutdown.UsecasesJobs++
	removeServiceEntity.gracefulShutdown.Unlock()
	defer decreaseJobs(removeServiceEntity.gracefulShutdown)
	// graceful shutdown part end

	logTryToGetCurrentServiceInfo(removeServiceName, removeServiceUUID, removeServiceEntity.logging)
	currentServiceInfo, err := removeServiceEntity.cacheStorage.GetServiceInfo(serviceInfo, removeServiceUUID)
	if err != nil {
		return fmt.Errorf("can't get current service info: %v", err)
	}
	logGotCurrentServiceInfo(removeServiceName, removeServiceUUID, currentServiceInfo, removeServiceEntity.logging)

	tunnelsFilesInfo := formTunnelsFilesInfo(currentServiceInfo.ApplicationServers, removeServiceEntity.cacheStorage)
	logTryCreateNewTunnels(removeServiceName, removeServiceUUID, tunnelsFilesInfo, removeServiceEntity.logging)
	oldTunnelsFilesInfo, err := removeServiceEntity.tunnelConfig.RemoveTunnels(tunnelsFilesInfo, removeServiceUUID)
	if err != nil {
		return fmt.Errorf("can't create tunnel files: %v", err)
	}
	logCreatedNewTunnels(removeServiceName, removeServiceUUID, tunnelsFilesInfo, removeServiceEntity.logging)

	// TODO: why double?
	if err := removeServiceEntity.cacheStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
		return fmt.Errorf("can't update tunnel info")
	}

	logTryRemoveIpvsadmService(removeServiceName, removeServiceUUID, currentServiceInfo, removeServiceEntity.logging)
	if err = removeServiceEntity.ipvsadm.RemoveService(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("ipvsadm can't remove service: %v. got error: %v", serviceInfo, err)
	}
	logRemovedIpvsadmService(removeServiceName, removeServiceUUID, currentServiceInfo, removeServiceEntity.logging)

	if err = removeServiceEntity.persistentStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return err
	}
	if err = removeServiceEntity.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return err
	}

	// TODO: why double?
	if err := removeServiceEntity.persistentStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
		return fmt.Errorf("can't update tunnel info")
	}

	logTryRemoveServiceAtHealtchecks(removeServiceName, removeServiceUUID, removeServiceEntity.logging)
	removeServiceEntity.hc.RemoveServiceFromHealtchecks(serviceInfo)
	logRemovedServiceAtHealtchecks(removeServiceName, removeServiceUUID, removeServiceEntity.logging)

	logTryRemoveIPFromDummy(removeServiceName, removeServiceUUID, serviceInfo.ServiceIP, removeServiceEntity.logging)
	if !removeServiceEntity.hc.isMockMode {
		if err = RemoveFromDummy(serviceInfo.ServiceIP); err != nil {
			return err
		}
	}
	logRemovedIPFromDummy(removeServiceName, removeServiceUUID, serviceInfo.ServiceIP, removeServiceEntity.logging)
	return nil
}
