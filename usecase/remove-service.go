package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const removeNlbServiceEntity = "remove-nlb-service"

// RemoveServiceEntity ...
type RemoveServiceEntity struct {
	locker            *domain.Locker
	configuratorVRRP  domain.ServiceWorker
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	gracefullShutdown *domain.GracefullShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewRemoveServiceEntity ...
func NewRemoveServiceEntity(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveServiceEntity {
	return &RemoveServiceEntity{
		locker:            locker,
		configuratorVRRP:  configuratorVRRP,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// RemoveService ...
// FIXME: rollbacks need refactor
func (removeServiceEntity *RemoveServiceEntity) RemoveService(serviceInfo domain.ServiceInfo,
	removeServiceUUID string) error {
	var err error

	// gracefull shutdown part start
	removeServiceEntity.locker.Lock()
	defer removeServiceEntity.locker.Unlock()
	removeServiceEntity.gracefullShutdown.Lock()
	if removeServiceEntity.gracefullShutdown.ShutdownNow {
		defer removeServiceEntity.gracefullShutdown.Unlock()
		return fmt.Errorf("program got shutdown signal, job remove service %v cancel", serviceInfo)
	}
	removeServiceEntity.gracefullShutdown.UsecasesJobs++
	removeServiceEntity.gracefullShutdown.Unlock()
	defer decreaseJobs(removeServiceEntity.gracefullShutdown)
	// gracefull shutdown part end

	currentServiceInfo, err := removeServiceEntity.cacheStorage.GetServiceInfo(serviceInfo, removeServiceUUID)
	if err = removeServiceEntity.configuratorVRRP.RemoveService(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("can't get current service info: %v", serviceInfo)
	}

	if err = removeServiceEntity.tunnelConfig.RemoveTunnels(currentServiceInfo.ApplicationServers, removeServiceUUID); err != nil {
		if errRollback := removeServiceEntity.cacheStorage.UpdateServiceInfo(serviceInfo, removeServiceUUID); errRollback != nil {
			// TODO: log it
		}
		return fmt.Errorf("can't remove tunnels: %v", err)
	}

	if err = removeServiceEntity.configuratorVRRP.RemoveService(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("configuratorVRRP can't remove service: %v", serviceInfo)
	}
	if err = removeServiceEntity.removeServiceFromPersistentStorage(serviceInfo, removeServiceUUID); err != nil {
		return err
	}
	if err = removeServiceEntity.removeNewServiceFromCacheStorage(serviceInfo, removeServiceUUID); err != nil {
		return err
	}
	return nil
}

func (removeServiceEntity *RemoveServiceEntity) removeServiceFromPersistentStorage(serviceInfo domain.ServiceInfo, removeServiceUUID string) error {
	if err := removeServiceEntity.persistentStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("error remove service %v from persistent storage: %v", serviceInfo, err)
	}
	return nil
}

func (removeServiceEntity *RemoveServiceEntity) removeNewServiceFromCacheStorage(serviceInfo domain.ServiceInfo, removeServiceUUID string) error {
	if err := removeServiceEntity.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("error remove service %v data from cache storage: %v", serviceInfo, err)
	}
	return nil
}
