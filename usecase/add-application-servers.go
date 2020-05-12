package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const addApplicationServersName = "add-application-servers"

// AddApplicationServers ...
type AddApplicationServers struct {
	locker            *domain.Locker
	ipvsadm           *portadapter.IPVSADMEntity // so dirty
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	hc                *HeathcheckEntity
	gracefullShutdown *domain.GracefullShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewAddApplicationServers ...
func NewAddApplicationServers(locker *domain.Locker,
	ipvsadm *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *AddApplicationServers {
	return &AddApplicationServers{
		locker:            locker,
		ipvsadm:           ipvsadm,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		gracefullShutdown: gracefullShutdown,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// AddNewApplicationServers ...
func (addApplicationServers *AddApplicationServers) AddNewApplicationServers(newServiceInfo *domain.ServiceInfo,
	addApplicationServersUUID string) (*domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo *domain.ServiceInfo

	// gracefull shutdown part start
	addApplicationServers.locker.Lock()
	defer addApplicationServers.locker.Unlock()
	addApplicationServers.gracefullShutdown.Lock()
	if addApplicationServers.gracefullShutdown.ShutdownNow {
		defer addApplicationServers.gracefullShutdown.Unlock()
		return newServiceInfo, fmt.Errorf("program got shutdown signal, job add application servers %v cancel", newServiceInfo)
	}
	addApplicationServers.gracefullShutdown.UsecasesJobs++
	addApplicationServers.gracefullShutdown.Unlock()
	defer decreaseJobs(addApplicationServers.gracefullShutdown)
	// gracefull shutdown part end

	// enrich application servers info start
	enrichedApplicationServers, err := addApplicationServers.tunnelConfig.EnrichApplicationServersInfo(newServiceInfo.ApplicationServers, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, err
	}
	newServiceInfo.ApplicationServers = enrichedApplicationServers
	// enrich application servers info end

	// need for rollback. used only service ip and port
	currentServiceInfo, err := addApplicationServers.cacheStorage.GetServiceInfo(newServiceInfo, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}
	updatedServiceInfo, err = forAddApplicationServersFormUpdateServiceInfo(currentServiceInfo, newServiceInfo, addApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't form update service info: %v", err)
	}
	// add to cache storage
	if err = addApplicationServers.updateServiceFromCacheStorage(updatedServiceInfo, addApplicationServersUUID); err != nil {
		if errRollBackCache := addApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, addApplicationServersUUID); errRollBackCache != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback cache, got error: %v", errRollBackCache)
		}
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}

	if err = addApplicationServers.tunnelConfig.CreateTunnels(enrichedApplicationServers, addApplicationServersUUID); err != nil {
		if errRollBackCache := addApplicationServers.cacheStorage.RemoveServiceDataFromStorage(newServiceInfo, addApplicationServersUUID); err != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback cache, got error: %v", errRollBackCache)
		}
		return updatedServiceInfo, err
	}

	if err = addApplicationServers.ipvsadm.AddApplicationServersForService(newServiceInfo, addApplicationServersUUID); err != nil {
		if errRollBackTunnels := addApplicationServers.tunnelConfig.RemoveTunnels(enrichedApplicationServers, addApplicationServersUUID); err != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackTunnels)
		}
		if errRollBackCache := addApplicationServers.cacheStorage.RemoveServiceDataFromStorage(newServiceInfo, addApplicationServersUUID); err != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback cache, got error: %v", errRollBackCache)
		}
		return currentServiceInfo, fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	if err = addApplicationServers.updateServiceFromPersistentStorage(updatedServiceInfo, addApplicationServersUUID); err != nil {
		if errRollBackTunnels := addApplicationServers.tunnelConfig.RemoveTunnels(enrichedApplicationServers, addApplicationServersUUID); err != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackTunnels)
		}

		if errRollBackCache := addApplicationServers.updateServiceFromCacheStorage(currentServiceInfo, addApplicationServersUUID); errRollBackCache != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback cache, got error: %v", errRollBackCache)
		}

		// TODO: when relese remove application service logic
		if errRollBackIPSVSADM := addApplicationServers.ipvsadm.RemoveApplicationServersFromService(newServiceInfo, addApplicationServersUUID); errRollBackIPSVSADM != nil {
			addApplicationServers.logging.WithFields(logrus.Fields{
				"entity":     addApplicationServersName,
				"event uuid": addApplicationServersUUID,
			}).Errorf("can't rollback ipvsadm, got error: %v", errRollBackIPSVSADM)
		}
		return currentServiceInfo, fmt.Errorf("Error when update persistent storage: %v", err)
	}
	addApplicationServers.hc.UpdateServiceAtHealtchecks(updatedServiceInfo)
	return updatedServiceInfo, nil
}

func (addApplicationServers *AddApplicationServers) updateServiceFromCacheStorage(serviceInfo *domain.ServiceInfo, addApplicationServersUUID string) error {
	if err := addApplicationServers.cacheStorage.UpdateServiceInfo(serviceInfo, addApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to cache storage: %v", err)
	}
	return nil
}

func (addApplicationServers *AddApplicationServers) updateServiceFromPersistentStorage(serviceInfo *domain.ServiceInfo, addApplicationServersUUID string) error {
	if err := addApplicationServers.persistentStorage.UpdateServiceInfo(serviceInfo, addApplicationServersUUID); err != nil {
		return fmt.Errorf("error add new service data to persistent storage: %v", err)
	}
	return nil
}
