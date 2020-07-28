package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const modifyServiceName = "modify service"

// ModifyServiceEntity ...
type ModifyServiceEntity struct {
	locker            *domain.Locker
	ipvsadm           *portadapter.IPVSADMEntity
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	tunnelConfig      domain.TunnelMaker
	hc                *HeathcheckEntity
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
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
	gracefulShutdown *domain.GracefulShutdown,
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
		gracefulShutdown:  gracefulShutdown,
		logging:           logging,
		uuidGenerator:     uuidGenerator,
	}
}

// ModifyService ...
func (modifyService *ModifyServiceEntity) ModifyService(serviceInfo *domain.ServiceInfo,
	modifyServiceUUID string) (*domain.ServiceInfo, error) {
	// graceful shutdown part start
	modifyService.locker.Lock()
	defer modifyService.locker.Unlock()
	modifyService.gracefulShutdown.Lock()
	if modifyService.gracefulShutdown.ShutdownNow {
		defer modifyService.gracefulShutdown.Unlock()
		return serviceInfo, fmt.Errorf("program got shutdown signal, job create service %v cancel", serviceInfo)
	}
	modifyService.gracefulShutdown.UsecasesJobs++
	modifyService.gracefulShutdown.Unlock()
	defer decreaseJobs(modifyService.gracefulShutdown)
	// graceful shutdown part end
	logTryToGetCurrentServiceInfo(modifyServiceName, modifyServiceUUID, modifyService.logging)
	currentServiceInfo, err := modifyService.cacheStorage.GetServiceInfo(serviceInfo, modifyServiceUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't get current service info: %v", err)
	}
	logGotCurrentServiceInfo(modifyServiceName, modifyServiceUUID, currentServiceInfo, modifyService.logging)

	logTryValidateForModifyService(modifyServiceName, modifyServiceUUID, modifyService.logging)
	if !modifyService.isServicesIPsAndPortsEqual(serviceInfo, currentServiceInfo, modifyServiceUUID) {
		return serviceInfo, fmt.Errorf("service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}
	logValidModifyService(modifyServiceName, modifyServiceUUID, modifyService.logging)

	// TODO: validate changes. ipvs, or only healtchecks
	logUpdateServiceAtHealtchecks(modifyServiceName, modifyServiceUUID, modifyService.logging)
	modifyService.hc.UpdateServiceAtHealtchecks(serviceInfo)
	logUpdatedServiceAtHealtchecks(modifyServiceName, modifyServiceUUID, modifyService.logging)

	logTryUpdateServiceInfoAtCache(modifyServiceName, modifyServiceUUID, modifyService.logging)
	if err = modifyService.cacheStorage.UpdateServiceInfo(serviceInfo, modifyServiceUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}
	logUpdateServiceInfoAtCache(modifyServiceName, modifyServiceUUID, modifyService.logging)

	logTryUpdateServiceInfoAtPersistentStorage(modifyServiceName, modifyServiceUUID, modifyService.logging)
	if err = modifyService.persistentStorage.UpdateServiceInfo(serviceInfo, modifyServiceUUID); err != nil {
		return currentServiceInfo, fmt.Errorf("error when update persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(modifyServiceName, modifyServiceUUID, modifyService.logging)

	return serviceInfo, nil
}

func (modifyService *ModifyServiceEntity) isServicesIPsAndPortsEqual(serviceOne,
	serviceTwo *domain.ServiceInfo, uuid string) bool {
	if serviceOne.ServiceIP != serviceTwo.ServiceIP ||
		serviceOne.ServicePort != serviceTwo.ServicePort {
		logServicesIPAndPortNotEqual(serviceOne.ServiceIP,
			serviceOne.ServicePort,
			serviceTwo.ServiceIP,
			serviceTwo.ServicePort,
			modifyServiceName,
			uuid,
			modifyService.logging)
		return false
	}
	if len(serviceOne.ApplicationServers) != len(serviceTwo.ServiceIP) {
		logServicesHaveDifferentNumberOfApplicationServers(serviceOne.ServiceIP, serviceOne.ServicePort, serviceTwo.ServiceIP, serviceTwo.ServicePort, len(serviceOne.ApplicationServers), len(serviceTwo.ServiceIP), modifyServiceName, uuid, modifyService.logging)
		return false
	}

	for _, applicationServerFromServiceOne := range serviceOne.ApplicationServers {
		var isFunded bool
		for _, applicationServerFromServiceTwo := range serviceTwo.ApplicationServers {
			if applicationServerFromServiceOne.ServerIP == applicationServerFromServiceTwo.ServerIP &&
				applicationServerFromServiceOne.ServerPort == applicationServerFromServiceTwo.ServerPort {
				isFunded = true
			}
		}
		if !isFunded {
			logApplicationServerNotFound(serviceOne.ServiceIP, serviceOne.ServicePort, applicationServerFromServiceOne.ServerIP, applicationServerFromServiceOne.ServerPort, modifyServiceName, uuid, modifyService.logging)
			return false
		}
	}
	return true
}
