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
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewRemoveServiceEntity ...
func NewRemoveServiceEntity(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
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
func (removeServiceEntity *RemoveServiceEntity) RemoveService(serviceInfo domain.ServiceInfo,
	removeServiceUUID string) error {
	var err error
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
	removeServiceEntity.persistentStorage.Lock()
	defer removeServiceEntity.persistentStorage.Unlock()
	if err := removeServiceEntity.persistentStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("error remove service %v from persistent storage: %v", serviceInfo, err)
	}
	return nil
}

func (removeServiceEntity *RemoveServiceEntity) removeNewServiceFromCacheStorage(serviceInfo domain.ServiceInfo, removeServiceUUID string) error {
	removeServiceEntity.cacheStorage.Lock()
	defer removeServiceEntity.cacheStorage.Unlock()
	if err := removeServiceEntity.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("error remove service %v data from cache storage: %v", serviceInfo, err)
	}
	return nil
}
