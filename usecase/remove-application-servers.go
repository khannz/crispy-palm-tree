package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const removeApplicationServersName = "remove-application-servers"

// RemoveApplicationServers ...
type RemoveApplicationServers struct {
	locker            *domain.Locker
	ipvsadm           *portadapter.IPVSADMEntity
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
	ipvsadm *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveApplicationServers {
	return &RemoveApplicationServers{
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
	tunnelsFilesInfo := formTunnelsFilesInfo(removeServiceInfo.ApplicationServers, removeApplicationServers.cacheStorage)
	oldTunnelsFilesInfo, err := removeApplicationServers.tunnelConfig.RemoveTunnels(tunnelsFilesInfo, removeApplicationServersUUID)
	if err != nil {
		return nil, fmt.Errorf("can't create tunnel files: %v", err)
	}

	// need for rollback. used only service ip and port
	currentServiceInfo, err := removeApplicationServers.cacheStorage.GetServiceInfo(removeServiceInfo, removeApplicationServersUUID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}

	if err = validateRemoveApplicationServers(currentServiceInfo.ApplicationServers, removeServiceInfo.ApplicationServers); err != nil {
		return updatedServiceInfo, fmt.Errorf("validate remove application servers fail: %v", err)
	}

	updatedServiceInfo = forRemoveApplicationServersFormUpdateServiceInfo(currentServiceInfo, removeServiceInfo, removeApplicationServersUUID) // ignore check unique error
	// update for cache storage
	if err = removeApplicationServers.cacheStorage.UpdateServiceInfo(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}

	if err = removeApplicationServers.cacheStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
		return currentServiceInfo, fmt.Errorf("can't update tunnel info in storage: %v", err)
	}

	if err = removeApplicationServers.ipvsadm.RemoveApplicationServersFromService(removeServiceInfo, removeApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	if err = removeApplicationServers.persistentStorage.UpdateServiceInfo(updatedServiceInfo, removeApplicationServersUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("Error when update persistent storage: %v", err)
	}
	if err = removeApplicationServers.persistentStorage.UpdateTunnelFilesInfoAtStorage(oldTunnelsFilesInfo); err != nil {
		return currentServiceInfo, fmt.Errorf("can't update tunnel info in storage: %v", err)
	}
	go removeApplicationServers.hc.UpdateServiceAtHealtchecks(updatedServiceInfo)
	return updatedServiceInfo, nil
}
