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

// NewCreateServiceEntity ...
func NewCreateServiceEntity(locker *domain.Locker,
	ipvsadm *portadapter.IPVSADMEntity,
	cacheStorage *portadapter.StorageEntity, // so dirty
	persistentStorage *portadapter.StorageEntity, // so dirty
	tunnelConfig domain.TunnelMaker,
	hc *HeathcheckEntity,
	commandGenerator domain.CommandGenerator,
	gracefullShutdown *domain.GracefullShutdown,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *CreateServiceEntity {
	return &CreateServiceEntity{
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

// CreateService ...
func (createService *CreateServiceEntity) CreateService(serviceInfo *domain.ServiceInfo,
	createServiceUUID string) (*domain.ServiceInfo, error) {
	// gracefull shutdown part start
	createService.locker.Lock()
	defer createService.locker.Unlock()
	createService.gracefullShutdown.Lock()
	if createService.gracefullShutdown.ShutdownNow {
		defer createService.gracefullShutdown.Unlock()
		return serviceInfo, fmt.Errorf("program got shutdown signal, job create service %v cancel", serviceInfo)
	}
	createService.gracefullShutdown.UsecasesJobs++
	createService.gracefullShutdown.Unlock()
	defer decreaseJobs(createService.gracefullShutdown)
	// gracefull shutdown part end

	tunnelsFilesInfo := formTunnelsFilesInfo(serviceInfo.ApplicationServers, createService.cacheStorage)

	newTunnelsFilesInfo, err := createService.tunnelConfig.CreateTunnels(tunnelsFilesInfo, createServiceUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't create tunnel files: %v", err)
	}
	// add to cache storage
	if err := createService.cacheStorage.NewServiceDataToStorage(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}
	if err := createService.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}

	if err := createService.ipvsadm.CreateService(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("Error when Configure VRRP: %v", err)
	}

	if err = createService.persistentStorage.NewServiceDataToStorage(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("Error when save to persistent storage: %v", err)
	}
	if err := createService.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
		return serviceInfo, fmt.Errorf("can't add to cache storage :%v", err)
	}

	if err := createService.commandGenerator.GenerateCommandsForApplicationServers(serviceInfo, createServiceUUID); err != nil {
		return serviceInfo, fmt.Errorf("can't generate commands :%v", err)
	}

	createService.hc.NewServiceToHealtchecks(serviceInfo)
	return serviceInfo, nil
}
