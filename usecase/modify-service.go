package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const modifyServiceName = "modify-service"

// ModifyServiceEntity ...
type ModifyServiceEntity struct {
	locker            *domain.Locker
	ipvsadm           *portadapter.IPVSADMEntity
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	hc                *HeathcheckEntity
	commandGenerator  domain.CommandGenerator
	gracefullShutdown *domain.GracefullShutdown
	uuidGenerator     domain.UUIDgenerator
	logging           *logrus.Logger
}

// NewModifyServiceEntity ...
func NewModifyServiceEntity(locker *domain.Locker,
	ipvsadm *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	commandGenerator domain.CommandGenerator,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *ModifyServiceEntity {
	return &ModifyServiceEntity{
		locker:            locker,
		ipvsadm:           ipvsadm,
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		tunnelConfig:      tunnelConfig,
		hc:                hc,
		commandGenerator:  commandGenerator,
		gracefullShutdown: gracefullShutdown,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// ModifyService ...
func (modifyService *ModifyServiceEntity) ModifyService(serviceInfo *domain.ServiceInfo,
	modifyServiceUUID string) (*domain.ServiceInfo, error) {
	// gracefull shutdown part start
	modifyService.locker.Lock()
	defer modifyService.locker.Unlock()
	modifyService.gracefullShutdown.Lock()
	if modifyService.gracefullShutdown.ShutdownNow {
		defer modifyService.gracefullShutdown.Unlock()
		return serviceInfo, fmt.Errorf("program got shutdown signal, job create service %v cancel", serviceInfo)
	}
	modifyService.gracefullShutdown.UsecasesJobs++
	modifyService.gracefullShutdown.Unlock()
	defer decreaseJobs(modifyService.gracefullShutdown)
	// gracefull shutdown part end

	// tunnelsFilesInfo := formTunnelsFilesInfo(serviceInfo.ApplicationServers, modifyService.cacheStorage)

	// newTunnelsFilesInfo, err := modifyService.tunnelConfig.CreateTunnels(tunnelsFilesInfo, modifyServiceUUID)
	// if err != nil {
	// 	return serviceInfo, fmt.Errorf("can't create tunnel files: %v", err)
	// }
	// // add to cache storage
	// if err := modifyService.cacheStorage.NewServiceDataToStorage(serviceInfo, modifyServiceUUID); err != nil {
	// 	return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	// }
	// if err := modifyService.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
	// 	return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	// }

	// if err := modifyService.ipvsadm.ModifyService(serviceInfo, modifyServiceUUID); err != nil {
	// 	return serviceInfo, fmt.Errorf("Error when Configure VRRP: %v", err)
	// }

	// if err = modifyService.persistentStorage.NewServiceDataToStorage(serviceInfo, modifyServiceUUID); err != nil {
	// 	return serviceInfo, fmt.Errorf("Error when save to persistent storage: %v", err)
	// }
	// if err := modifyService.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
	// 	return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	// }

	// if err := modifyService.commandGenerator.GenerateCommandsForApplicationServers(serviceInfo, modifyServiceUUID); err != nil {
	// 	return serviceInfo, fmt.Errorf("can't generate commands :%v", err)
	// }

	// modifyService.hc.NewServiceToHealtchecks(serviceInfo)
	return serviceInfo, nil
}
