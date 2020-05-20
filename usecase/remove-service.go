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
	ipvsadm           *portadapter.IPVSADMEntity
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	gracefullShutdown *domain.GracefullShutdown
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
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	hc *HeathcheckEntity,
	logging *logrus.Logger) *RemoveServiceEntity {
	return &RemoveServiceEntity{
		locker:            locker,
		ipvsadm:           ipvsadm,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		gracefullShutdown: gracefullShutdown,
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
	if err != nil {
		return fmt.Errorf("can't get current service info: %v", err)
	}

	tunnelsFilesInfo := formTunnelsFilesInfo(currentServiceInfo.ApplicationServers, removeServiceEntity.cacheStorage)
	oldTunnelsFilesInfo, err := removeServiceEntity.tunnelConfig.RemoveTunnels(tunnelsFilesInfo, removeServiceUUID)
	if err != nil {
		return fmt.Errorf("can't create tunnel files: %v", err)
	}
	if err := removeServiceEntity.cacheStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
		return fmt.Errorf("can't update tunnel info")
	}

	if err = removeServiceEntity.ipvsadm.RemoveService(serviceInfo, removeServiceUUID); err != nil {
		return fmt.Errorf("ipvsadm can't remove service: %v. got error: %v", serviceInfo, err)
	}
	if err = removeServiceEntity.persistentStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return err
	}
	if err = removeServiceEntity.cacheStorage.RemoveServiceDataFromStorage(serviceInfo, removeServiceUUID); err != nil {
		return err
	}

	if err := removeServiceEntity.persistentStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
		return fmt.Errorf("can't update tunnel info")
	}

	go removeServiceEntity.hc.RemoveServiceFromHealtchecks(serviceInfo)

	if !removeServiceEntity.hc.isMockMode {
		if err = RemoveFromDummy(serviceInfo.ServiceIP); err != nil {
			return err
		}
	}

	return nil
}
