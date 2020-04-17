package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const removeApplicationServers = "remove-application-servers"

// RemoveApplicationServers ...
type RemoveApplicationServers struct {
	locker            *domain.Locker
	configuratorVRRP  domain.ServiceWorker
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewRemoveApplicationServers ...
func NewRemoveApplicationServers(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveApplicationServers {
	return &RemoveApplicationServers{
		locker:            locker,
		configuratorVRRP:  configuratorVRRP,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// AddApplicationServers ...
func (removeApplicationServers *RemoveApplicationServers) AddApplicationServers(removeServiceInfo domain.ServiceInfo,
	removeApplicationServersUUID string) (domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo domain.ServiceInfo
	// deployedEntities := map[string][]string{}
	// deployedEntities, err = removeApplicationServers.tunnelConfig.CreateTunnel(deployedEntities, applicationServers, removeApplicationServersUUID)
	// if err != nil {
	// 	tunnelsRemove(deployedEntities, removeApplicationServers.tunnelConfig, removeApplicationServersUUID)
	// 	return updatedServiceInfo, fmt.Errorf("Error when create tunnel: %v", err)
	// }

	// need for rollback. used only service ip and port
	currentServiceInfo, err := removeApplicationServers.getServiceInfo(removeServiceInfo, removeApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}

	if err = validateRemoveApplicationServers(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers); err != nil {
		return updatedServiceInfo, fmt.Errorf("validate remove application servers fail: %v", err)
	}

	updatedServiceInfo = removeApplicationServers.formUpdateServiceInfo(currentServiceInfo, removeServiceInfo, removeApplicationServersUUID)

	// add to cache storage
	if err = removeApplicationServers.updateServiceFromCacheStorage(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		if errRollback := removeApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, removeApplicationServersUUID); errRollback != nil {
			// TODO: log it
		}
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}

	if err = removeApplicationServers.configuratorVRRP.RemoveApplicationServersFromService(removeServiceInfo, removeApplicationServersUUID); err != nil {
		if errRollback := removeApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, removeApplicationServersUUID); errRollback != nil {
			// TODO: log it
		}
		return currentServiceInfo, fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	if err = removeApplicationServers.updateServiceFromPersistentStorage(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		if errRollBackCache := removeApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, removeApplicationServersUUID); errRollBackCache != nil {
			// TODO: log: cant roll back
		}

		if errRollBackCache := removeApplicationServers.configuratorVRRP.AddApplicationServersFromService(removeServiceInfo, removeApplicationServersUUID); errRollBackCache != nil {
			// TODO: log: cant roll back
		}
		return currentServiceInfo, fmt.Errorf("Error when update persistent storage: %v", err)
	}

	return updatedServiceInfo, nil
}

func (removeApplicationServers *RemoveApplicationServers) updateServiceFromCacheStorage(serviceInfo domain.ServiceInfo, removeApplicationServersUUID string) error {
	removeApplicationServers.cacheStorage.Lock()
	defer removeApplicationServers.cacheStorage.Unlock()
	if err := removeApplicationServers.cacheStorage.UpdateServiceInfo(serviceInfo, removeApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to cache storage: %v", err)
	}
	return nil
}

func (removeApplicationServers *RemoveApplicationServers) updateServiceFromPersistentStorage(serviceInfo domain.ServiceInfo, removeApplicationServersUUID string) error {
	if err := removeApplicationServers.persistentStorage.UpdateServiceInfo(serviceInfo, removeApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to persistent storage: %v", err)
	}
	return nil
}

func (removeApplicationServers *RemoveApplicationServers) getServiceInfo(removeServiceInfo domain.ServiceInfo,
	removeApplicationServersUUID string) (domain.ServiceInfo, error) {
	removeApplicationServers.cacheStorage.Lock()
	defer removeApplicationServers.cacheStorage.Unlock()
	return removeApplicationServers.cacheStorage.GetServiceInfo(removeServiceInfo, removeApplicationServersUUID)
}

func (removeApplicationServers *RemoveApplicationServers) formUpdateServiceInfo(currentServiceInfo, removeServiceInfo domain.ServiceInfo, eventUUID string) domain.ServiceInfo {
	var resultServiceInfo domain.ServiceInfo
	resultApplicationServers := formNewApplicationServersSlice(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers)
	resultServiceInfo = domain.ServiceInfo{
		ServiceIP:          removeServiceInfo.ServiceIP,
		ServicePort:        removeServiceInfo.ServicePort,
		ApplicationServers: resultApplicationServers,
	}
	return resultServiceInfo
}
