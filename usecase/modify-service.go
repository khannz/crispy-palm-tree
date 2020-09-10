package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const modifyServiceName = "modify service"

// ModifyServiceEntity ...
type ModifyServiceEntity struct {
	locker            *domain.Locker
	ipvsadm           domain.IPVSWorker
	cacheStorage      domain.StorageActions
	persistentStorage domain.StorageActions
	tunnelConfig      domain.TunnelMaker
	hc                domain.HeathcheckWorker
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
	logging           *logrus.Logger
}

// NewModifyServiceEntity ...
func NewModifyServiceEntity(locker *domain.Locker,
	ipvsadm domain.IPVSWorker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc domain.HeathcheckWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
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
	logTryPreValidateRequest(modifyServiceName, modifyServiceUUID, modifyService.logging)
	allCurrentServices, err := modifyService.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return serviceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}

	if !isServiceExist(serviceInfo.ServiceIP, serviceInfo.ServicePort, allCurrentServices) {
		return serviceInfo, fmt.Errorf("service %v:%v does not exist, can't modify", serviceInfo.ServiceIP, serviceInfo.ServicePort)
	}

	logTryToGetCurrentServiceInfo(modifyServiceName, modifyServiceUUID, modifyService.logging)
	currentServiceInfo, err := modifyService.cacheStorage.GetServiceInfo(serviceInfo, modifyServiceUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't get current service info: %v", err)
	}
	logGotCurrentServiceInfo(modifyServiceName, modifyServiceUUID, currentServiceInfo, modifyService.logging)

	if err = checkApplicationServersExistInService(serviceInfo.ApplicationServers, currentServiceInfo); err != nil {
		return serviceInfo, err
	}

	if err = checkRoutingTypeForApplicationServersValid(serviceInfo, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	if !modifyService.isServicesIPsAndPortsEqual(serviceInfo, currentServiceInfo, modifyServiceUUID) {
		return serviceInfo, fmt.Errorf("service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}

	if serviceInfo.RoutingType != currentServiceInfo.RoutingType {
		return serviceInfo, fmt.Errorf("routing type at service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}
	if serviceInfo.Protocol != currentServiceInfo.Protocol {
		return serviceInfo, fmt.Errorf("protocol at service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}
	logPreValidateRequestIsOk(modifyServiceName, modifyServiceUUID, modifyService.logging)

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

	logUpdateServiceAtHealtchecks(modifyServiceName, modifyServiceUUID, modifyService.logging)
	if err = modifyService.hc.UpdateServiceAtHealtchecks(serviceInfo); err != nil {
		return serviceInfo, fmt.Errorf("service modify for healtchecks not activated, an error occurred when changing healtchecks: %v", err)
	}
	logUpdatedServiceAtHealtchecks(modifyServiceName, modifyServiceUUID, modifyService.logging)

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
	if len(serviceOne.ApplicationServers) != len(serviceTwo.ApplicationServers) {
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
