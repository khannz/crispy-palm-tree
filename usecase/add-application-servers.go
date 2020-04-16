package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const addApplicationServers = "add-application-servers"

// AddApplicationServers ...
type AddApplicationServers struct {
	locker            *domain.Locker
	configuratorVRRP  domain.ServiceWorker
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewAddApplicationServers ...
func NewAddApplicationServers(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *AddApplicationServers {
	return &AddApplicationServers{
		locker:            locker,
		configuratorVRRP:  configuratorVRRP,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// AddNewApplicationServers ...
func (addApplicationServers *AddApplicationServers) AddNewApplicationServers(newServiceInfo domain.ServiceInfo,
	addApplicationServersUUID string) (domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo domain.ServiceInfo
	// deployedEntities := map[string][]string{}
	// deployedEntities, err = addApplicationServers.tunnelConfig.CreateTunnel(deployedEntities, applicationServers, addApplicationServersUUID)
	// if err != nil {
	// 	tunnelsRemove(deployedEntities, addApplicationServers.tunnelConfig, addApplicationServersUUID)
	// 	return updatedServiceInfo, fmt.Errorf("Error when create tunnel: %v", err)
	// }

	// need for rollback. used only service ip and port
	currentServiceInfo, err := addApplicationServers.getServiceInfo(newServiceInfo, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}

	updatedServiceInfo, err = addApplicationServers.formUpdateServiceInfo(currentServiceInfo, newServiceInfo, addApplicationServersUUID)
	// add to cache storage
	if err = addApplicationServers.updateServiceFromCacheStorage(updatedServiceInfo, addApplicationServersUUID); err != nil {
		if errRollback := addApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, addApplicationServersUUID); errRollback != nil {
			// TODO: log it
		}
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}

	if err = addApplicationServers.configuratorVRRP.UpdateService(newServiceInfo, addApplicationServersUUID); err != nil {
		if errRollback := addApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, addApplicationServersUUID); errRollback != nil {
			// TODO: log it
		}
		return currentServiceInfo, fmt.Errorf("can't update service: %v", err)
	}

	if err = addApplicationServers.updateServiceFromPersistentStorage(updatedServiceInfo, addApplicationServersUUID); err != nil {
		if errRollBackCache := addApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, addApplicationServersUUID); errRollBackCache != nil {
			// TODO: log: cant roll back
		}

		// TODO: when relese remove application service logic
		// if errRollBackCache := addApplicationServers.configuratorVRRP.RemoveApplicationServersFromService(newServiceInfo, addApplicationServersUUID); errRollBackCache != nil {
		// 	// TODO: log: cant roll back
		// }
		return currentServiceInfo, fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	return updatedServiceInfo, nil
}

func (addApplicationServers *AddApplicationServers) updateServiceFromCacheStorage(serviceInfo domain.ServiceInfo, addApplicationServersUUID string) error {
	addApplicationServers.cacheStorage.Lock()
	defer addApplicationServers.cacheStorage.Unlock()
	if err := addApplicationServers.cacheStorage.UpdateServiceInfo(serviceInfo, addApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to cache storage: %v", err)
	}
	return nil
}

func (addApplicationServers *AddApplicationServers) updateServiceFromPersistentStorage(serviceInfo domain.ServiceInfo, addApplicationServersUUID string) error {
	if err := addApplicationServers.persistentStorage.UpdateServiceInfo(serviceInfo, addApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to persistent storage: %v", err)
	}
	return nil
}

func (addApplicationServers *AddApplicationServers) getServiceInfo(newServiceInfo domain.ServiceInfo,
	addApplicationServersUUID string) (domain.ServiceInfo, error) {
	addApplicationServers.cacheStorage.Lock()
	defer addApplicationServers.cacheStorage.Unlock()
	return addApplicationServers.cacheStorage.GetServiceInfo(newServiceInfo, addApplicationServersUUID)
}

func (addApplicationServers *AddApplicationServers) formUpdateServiceInfo(currentServiceInfo, newServiceInfo domain.ServiceInfo, eventUUID string) (domain.ServiceInfo, error) {
	var resultServiceInfo domain.ServiceInfo
	if err := checkNewApplicationServersIsUnique(currentServiceInfo, newServiceInfo, eventUUID); err != nil {
		return resultServiceInfo, fmt.Errorf("new application server not unique: %v", err)
	}
	// concatenate two slices
	resultApplicationServers := append(currentServiceInfo.ApplicationServers, newServiceInfo.ApplicationServers...)

	resultServiceInfo = domain.ServiceInfo{
		ServiceIP:          newServiceInfo.ServiceIP,
		ServicePort:        newServiceInfo.ServicePort,
		ApplicationServers: resultApplicationServers,
	}
	return resultServiceInfo, nil
}

// TODO: need better check unique, app srv to services too
func checkNewApplicationServersIsUnique(currentServiceInfo, newServiceInfo domain.ServiceInfo, eventUUID string) error {
	// TODO: bad loops
	for _, newApplicationServer := range newServiceInfo.ApplicationServers {
		for _, currentApplicationServer := range currentServiceInfo.ApplicationServers {
			if newApplicationServer == currentApplicationServer {
				return fmt.Errorf("application server %v:%v alredy exist in service %v:%v",
					newApplicationServer.ServerIP,
					newApplicationServer.ServerPort,
					newServiceInfo.ServiceIP,
					newServiceInfo.ServicePort)
			}
		}
	}
	return nil
}
