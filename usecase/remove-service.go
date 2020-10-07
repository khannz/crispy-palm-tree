package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/healthcheck"
	"github.com/sirupsen/logrus"
)

const removeServiceName = "remove-service"

// RemoveServiceEntity ...
type RemoveServiceEntity struct {
	locker            *domain.Locker
	cacheStorage      domain.StorageActions
	persistentStorage domain.StorageActions
	tunnelConfig      domain.TunnelMaker
	hc                *healthcheck.HeathcheckEntity
	gracefulShutdown  *domain.GracefulShutdown
	logging           *logrus.Logger
}

// NewRemoveServiceEntity ...
func NewRemoveServiceEntity(locker *domain.Locker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *RemoveServiceEntity {
	return &RemoveServiceEntity{
		locker:            locker,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		gracefulShutdown:  gracefulShutdown,
		logging:           logging,
	}
}

// RemoveService ...
// TODO: rollbacks need refactor
func (removeServiceEntity *RemoveServiceEntity) RemoveService(serviceInfo *domain.ServiceInfo,
	removeServiceID string) error {
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
	logStartUsecase(removeServiceName, "remove service", removeServiceID, serviceInfo, removeServiceEntity.logging)
	allCurrentServices, err := removeServiceEntity.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return fmt.Errorf("fail when loading info about current services: %v", err)
	}

	if !isServiceExist(serviceInfo.IP, serviceInfo.Port, allCurrentServices) {
		return fmt.Errorf("service %v:%v not exist, can't remove it", serviceInfo.IP, serviceInfo.Port)
	}

	logTryToGetCurrentServiceInfo(removeServiceName, removeServiceID, removeServiceEntity.logging)
	currentServiceInfo, err := removeServiceEntity.cacheStorage.GetServiceInfo(serviceInfo, removeServiceID)
	if err != nil {
		return fmt.Errorf("can't get current service info: %v", err)
	}
	logGotCurrentServiceInfo(removeServiceName, removeServiceID, currentServiceInfo, removeServiceEntity.logging)
	logTryPreValidateRequest(removeServiceName, removeServiceID, removeServiceEntity.logging)

	logTryRemoveServiceAtHealtchecks(removeServiceName, removeServiceID, removeServiceEntity.logging)
	hcService := healthcheck.ConvertDomainServiceToHCService(serviceInfo)
	removeServiceEntity.hc.RemoveServiceFromHealtchecks(hcService) // will wait until removed
	logRemovedServiceAtHealtchecks(removeServiceName, removeServiceID, removeServiceEntity.logging)

	var tunnelsFilesInfo, oldTunnelsFilesInfo []*domain.TunnelForApplicationServer
	if currentServiceInfo.Protocol == "tcp" {
		tunnelsFilesInfo = FormTunnelsFilesInfo(currentServiceInfo.ApplicationServers, removeServiceEntity.cacheStorage)
		logTryRemoveTunnels(removeServiceName, removeServiceID, tunnelsFilesInfo, removeServiceEntity.logging)
		oldTunnelsFilesInfo, err = removeServiceEntity.tunnelConfig.RemoveTunnels(tunnelsFilesInfo, removeServiceID)
		if err != nil {
			return fmt.Errorf("can't remove tunnel files: %v", err)
		}
		logRemovedTunnels(removeServiceName, removeServiceID, tunnelsFilesInfo, removeServiceEntity.logging)

		if err := removeServiceEntity.cacheStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
			return fmt.Errorf("can't update tunnel info")
		}
	}

	removeServiceEntity.logging.WithFields(logrus.Fields{
		"event id": removeServiceID,
	}).Debugf("try remove from storages service %v", serviceInfo)
	if err = removeServiceEntity.persistentStorage.RemoveServiceInfoFromStorage(serviceInfo, removeServiceID); err != nil {
		return err
	}
	if err = removeServiceEntity.cacheStorage.RemoveServiceInfoFromStorage(serviceInfo, removeServiceID); err != nil {
		return err
	}
	removeServiceEntity.logging.WithFields(logrus.Fields{
		"event id": removeServiceID,
	}).Debugf("removed from storages service %v", serviceInfo)

	if currentServiceInfo.Protocol == "tcp" {
		if err := removeServiceEntity.persistentStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
			return fmt.Errorf("can't update tunnel info")
		}
	}

	return nil
}
