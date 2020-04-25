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
	hc                *HeathcheckEntity
	gracefullShutdown *domain.GracefullShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewRemoveApplicationServers ...
func NewRemoveApplicationServers(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveApplicationServers {
	return &RemoveApplicationServers{
		locker:            locker,
		configuratorVRRP:  configuratorVRRP,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		gracefullShutdown: gracefullShutdown,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// RemoveApplicationServers ...
// FIXME: rollbacks need refactor
func (removeApplicationServers *RemoveApplicationServers) RemoveApplicationServers(removeServiceInfo *domain.ServiceInfo,
	removeApplicationServersUUID string) (*domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo *domain.ServiceInfo

	// gracefull shutdown part start
	removeApplicationServers.locker.Lock()
	defer removeApplicationServers.locker.Unlock()
	removeApplicationServers.gracefullShutdown.Lock()
	if removeApplicationServers.gracefullShutdown.ShutdownNow {
		defer removeApplicationServers.gracefullShutdown.Unlock()
		return removeServiceInfo, fmt.Errorf("program got shutdown signal, job remove application servers %v cancel", removeServiceInfo)
	}
	removeApplicationServers.gracefullShutdown.UsecasesJobs++
	removeApplicationServers.gracefullShutdown.Unlock()
	defer decreaseJobs(removeApplicationServers.gracefullShutdown)
	// gracefull shutdown part end

	// need for rollback. used only service ip and port
	currentServiceInfo, err := removeApplicationServers.getServiceInfo(removeServiceInfo, removeApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}

	if err = validateRemoveApplicationServers(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers); err != nil {
		return updatedServiceInfo, fmt.Errorf("validate remove application servers fail: %v", err)
	}

	enrichedApplicationServersForRemove := enrichApplicationServersInfo(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers)

	updatedServiceInfo = removeApplicationServers.formUpdateServiceInfo(currentServiceInfo, removeServiceInfo, removeApplicationServersUUID)
	// add to cache storage
	if err = removeApplicationServers.updateServiceFromCacheStorage(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		if errRollback := removeApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, removeApplicationServersUUID); errRollback != nil {
			// TODO: log it
		}
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}

	if err = removeApplicationServers.tunnelConfig.RemoveTunnels(enrichedApplicationServersForRemove, removeApplicationServersUUID); err != nil {
		if errRollback := removeApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, removeApplicationServersUUID); errRollback != nil {
			// TODO: log it
		}
		return currentServiceInfo, fmt.Errorf("can't remove tunnels: %v", err)
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

		if errRollBackCache := removeApplicationServers.configuratorVRRP.AddApplicationServersForService(removeServiceInfo, removeApplicationServersUUID); errRollBackCache != nil {
			// TODO: log: cant roll back
		}
		return currentServiceInfo, fmt.Errorf("Error when update persistent storage: %v", err)
	}
	removeApplicationServers.hc.CheckApplicationServersInService(updatedServiceInfo)
	return updatedServiceInfo, nil
}

func (removeApplicationServers *RemoveApplicationServers) updateServiceFromCacheStorage(serviceInfo *domain.ServiceInfo, removeApplicationServersUUID string) error {
	if err := removeApplicationServers.cacheStorage.UpdateServiceInfo(serviceInfo, removeApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to cache storage: %v", err)
	}
	return nil
}

func (removeApplicationServers *RemoveApplicationServers) updateServiceFromPersistentStorage(serviceInfo *domain.ServiceInfo, removeApplicationServersUUID string) error {
	if err := removeApplicationServers.persistentStorage.UpdateServiceInfo(serviceInfo, removeApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to persistent storage: %v", err)
	}
	return nil
}

func (removeApplicationServers *RemoveApplicationServers) getServiceInfo(removeServiceInfo *domain.ServiceInfo,
	removeApplicationServersUUID string) (*domain.ServiceInfo, error) {
	return removeApplicationServers.cacheStorage.GetServiceInfo(removeServiceInfo, removeApplicationServersUUID)
}

func (removeApplicationServers *RemoveApplicationServers) formUpdateServiceInfo(currentServiceInfo, removeServiceInfo *domain.ServiceInfo, eventUUID string) *domain.ServiceInfo {
	var resultServiceInfo *domain.ServiceInfo
	resultApplicationServers := formNewApplicationServersSlice(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers)
	resultServiceInfo = &domain.ServiceInfo{
		ServiceIP:          removeServiceInfo.ServiceIP,
		ServicePort:        removeServiceInfo.ServicePort,
		ApplicationServers: resultApplicationServers,
	}
	return resultServiceInfo
}
