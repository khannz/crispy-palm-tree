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
	hc                *HeathcheckEntity
	gracefullShutdown *domain.GracefullShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewCreateServiceEntity ...
func NewCreateServiceEntity(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *CreateServiceEntity {
	return &CreateServiceEntity{
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

// CreateService ...
func (createService *CreateServiceEntity) CreateService(serviceInfo *domain.ServiceInfo,
	createServiceUUID string) error {
	// gracefull shutdown part start
	createService.locker.Lock()
	defer createService.locker.Unlock()
	createService.gracefullShutdown.Lock()
	if createService.gracefullShutdown.ShutdownNow {
		defer createService.gracefullShutdown.Unlock()
		return fmt.Errorf("program got shutdown signal, job create service %v cancel", serviceInfo)
	}
	createService.gracefullShutdown.UsecasesJobs++
	createService.gracefullShutdown.Unlock()
	defer decreaseJobs(createService.gracefullShutdown)
	// gracefull shutdown part end

	// enrich application servers info start
	enrichedApplicationServers, err := createService.tunnelConfig.EnrichApplicationServersInfo(serviceInfo.ApplicationServers, createServiceUUID)
	if err != nil {
		return err
	}
	serviceInfo.ApplicationServers = enrichedApplicationServers
	// enrich application servers info end

	// add to cache storage
	if err = createService.cacheStorage.NewServiceDataToStorage(serviceInfo, createServiceUUID); err != nil {
		return fmt.Errorf("can't add to cache storage :%v", err)
	}

	if err = createService.tunnelConfig.CreateTunnels(enrichedApplicationServers, createServiceUUID); err != nil {
		if errRollBackCache := createService.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback cache, got error: %v", errRollBackCache)
		}
		return err
	}

	if err = createService.configuratorVRRP.CreateService(serviceInfo, createServiceUUID); err != nil {
		if errRollBackTunnels := createService.tunnelConfig.RemoveTunnels(enrichedApplicationServers, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackTunnels)
		}
		if errRollBackCache := createService.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackCache)
		}
		if errRollBackIPVSADM := createService.configuratorVRRP.RemoveService(serviceInfo, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackIPVSADM)
		}

		return fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	if err = createService.persistentStorage.NewServiceDataToStorage(serviceInfo, createServiceUUID); err != nil {
		if errRollBackTunnels := createService.tunnelConfig.RemoveTunnels(enrichedApplicationServers, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackTunnels)
		}

		if errRollBackCache := createService.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackCache)
		}

		if errRollBackIPVSADM := createService.configuratorVRRP.RemoveService(serviceInfo, createServiceUUID); err != nil {
			createService.logging.WithFields(logrus.Fields{
				"entity":     createServiceName,
				"event uuid": createServiceUUID,
			}).Errorf("can't rollback tunnels, got error: %v", errRollBackIPVSADM)
		}

		return fmt.Errorf("Error when Configure VRRP: %v", err)
	}
	createService.hc.CheckApplicationServersInService(serviceInfo)
	return nil
}
