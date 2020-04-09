package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const createServiceName = "create-service"

// CreateServiceEntity ...
type CreateServiceEntity struct {
	locker            *domain.Locker
	configuratorVRRP  domain.ServiceWorker
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewCreateServiceEntity ...
func NewCreateServiceEntity(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *CreateServiceEntity {
	return &CreateServiceEntity{
		locker:            locker,
		configuratorVRRP:  configuratorVRRP,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// CreateService ...
func (createService *CreateServiceEntity) CreateService(serviceInfo domain.ServiceInfo, createServiceUUID string) error {
	var err error
	// deployedEntities := map[string][]string{}
	// deployedEntities, err = createService.tunnelConfig.CreateTunnel(deployedEntities, applicationServers, newNWBRequestUUID)
	// if err != nil {
	// 	tunnelsRemove(deployedEntities, createService.tunnelConfig, newNWBRequestUUID)
	// 	return fmt.Errorf("Error when create tunnel: %v", err)
	// }

	// add to cache storage
	if err = createService.addNewServiceToCacheStorage(serviceInfo, createServiceUUID); err != nil {
		return fmt.Errorf("can't add to cache storage :%v", err)
	}

	if err = createService.configuratorVRRP.CreateService(serviceInfo, createServiceUUID); err != nil {
		if errRollBackCache := createService.removeNewServiceFromCacheStorage(serviceInfo, createServiceUUID); errRollBackCache != nil {
			// TODO: log: cant roll back
		}
		return fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	if err = createService.addNewServiceToPersistentStorage(serviceInfo, createServiceUUID); err != nil {
		if errRollBackCache := createService.removeNewServiceFromCacheStorage(serviceInfo, createServiceUUID); errRollBackCache != nil {
			// TODO: log: cant roll back
		}

		// if errRollBackCache := createService.configuratorVRRP.RemoveService(serviceInfo, createServiceUUID); errRollBackCache != nil {
		// 	// TODO: log: cant roll back
		// }
		return fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	return nil
}

func (createService *CreateServiceEntity) addNewServiceToCacheStorage(serviceInfo domain.ServiceInfo, createServiceUUID string) error {
	createService.cacheStorage.Lock()
	defer createService.cacheStorage.Unlock()
	if err := createService.cacheStorage.NewServiceDataToStorage(serviceInfo, createServiceUUID); err != nil {
		return fmt.Errorf("Error add new service data to storage: %v", err)
	}
	return nil
}

func (createService *CreateServiceEntity) addNewServiceToPersistentStorage(serviceInfo domain.ServiceInfo, createServiceUUID string) error {
	if err := createService.persistentStorage.NewServiceDataToStorage(serviceInfo, createServiceUUID); err != nil {
		return fmt.Errorf("Error add new service data to storage: %v", err)
	}
	return nil
}

func (createService *CreateServiceEntity) removeNewServiceFromCacheStorage(serviceInfo domain.ServiceInfo, createServiceUUID string) error {
	createService.cacheStorage.Lock()
	defer createService.cacheStorage.Unlock()
	if err := createService.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, createServiceUUID); err != nil {
		return fmt.Errorf("Error add new service data to storage: %v", err)
	}
	return nil
}
