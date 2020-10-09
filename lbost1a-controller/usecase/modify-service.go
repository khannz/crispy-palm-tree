package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/lbost1a-controller/domain"
	"github.com/sirupsen/logrus"
)

const modifyServiceName = "modify service"

// ModifyServiceEntity ...
type ModifyServiceEntity struct {
	locker            *domain.Locker
	cacheStorage      domain.StorageActions
	persistentStorage domain.StorageActions
	tunnelConfig      domain.TunnelMaker
	hc                domain.HCWorker
	commandGenerator  domain.CommandGenerator
	gracefulShutdown  *domain.GracefulShutdown
	logging           *logrus.Logger
}

// NewModifyServiceEntity ...
func NewModifyServiceEntity(locker *domain.Locker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc domain.HCWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *ModifyServiceEntity {
	return &ModifyServiceEntity{
		locker:            locker,
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
	modifyServiceID string) (*domain.ServiceInfo, error) {
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
	logTryPreValidateRequest(modifyServiceName, modifyServiceID, modifyService.logging)
	allCurrentServices, err := modifyService.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return serviceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}

	if !isServiceExist(serviceInfo.IP, serviceInfo.Port, allCurrentServices) {
		return serviceInfo, fmt.Errorf("service %v:%v does not exist, can't modify", serviceInfo.IP, serviceInfo.Port)
	}

	logTryToGetCurrentServiceInfo(modifyServiceName, modifyServiceID, modifyService.logging)
	currentServiceInfo, err := modifyService.cacheStorage.GetServiceInfo(serviceInfo, modifyServiceID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't get current service info: %v", err)
	}
	logGotCurrentServiceInfo(modifyServiceName, modifyServiceID, currentServiceInfo, modifyService.logging)

	if err = checkApplicationServersExistInService(serviceInfo.ApplicationServers, currentServiceInfo); err != nil {
		return serviceInfo, err
	}

	if err = checkRoutingTypeForApplicationServersValid(serviceInfo, allCurrentServices); err != nil {
		return serviceInfo, err
	}

	// FIXME: check BalanceType!
	if !modifyService.isServicesIPsAndPortsEqual(serviceInfo, currentServiceInfo, modifyServiceID) {
		return serviceInfo, fmt.Errorf("service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}

	if serviceInfo.RoutingType != currentServiceInfo.RoutingType {
		return serviceInfo, fmt.Errorf("routing type at service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}
	if serviceInfo.Protocol != currentServiceInfo.Protocol {
		return serviceInfo, fmt.Errorf("protocol at service for modify and current service not equal, cannot modify: %v", currentServiceInfo)
	}
	logPreValidateRequestIsOk(modifyServiceName, modifyServiceID, modifyService.logging)

	logTryUpdateServiceInfoAtCache(modifyServiceName, modifyServiceID, modifyService.logging)
	if err = modifyService.cacheStorage.UpdateServiceInfo(serviceInfo, modifyServiceID); err != nil {
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}
	logUpdateServiceInfoAtCache(modifyServiceName, modifyServiceID, modifyService.logging)

	logTryUpdateServiceInfoAtPersistentStorage(modifyServiceName, modifyServiceID, modifyService.logging)
	if err = modifyService.persistentStorage.UpdateServiceInfo(serviceInfo, modifyServiceID); err != nil {
		return currentServiceInfo, fmt.Errorf("error when update persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(modifyServiceName, modifyServiceID, modifyService.logging)

	logUpdateServiceAtHealtchecks(modifyServiceName, modifyServiceID, modifyService.logging)
	hcServiceInfo, err := modifyService.hc.UpdateServiceAtHealtchecks(serviceInfo)
	if err != nil {
		return serviceInfo, fmt.Errorf("service modify for healtchecks not activated, an error occurred when changing healtchecks: %v", err)
	}
	logUpdatedServiceAtHealtchecks(modifyServiceName, modifyServiceID, modifyService.logging)

	return hcServiceInfo, nil
}

func (modifyService *ModifyServiceEntity) isServicesIPsAndPortsEqual(serviceOne,
	serviceTwo *domain.ServiceInfo, id string) bool {
	if serviceOne.IP != serviceTwo.IP ||
		serviceOne.Port != serviceTwo.Port {
		logServicesIPAndPortNotEqual(serviceOne.IP,
			serviceOne.Port,
			serviceTwo.IP,
			serviceTwo.Port,
			modifyServiceName,
			id,
			modifyService.logging)
		return false
	}
	if len(serviceOne.ApplicationServers) != len(serviceTwo.ApplicationServers) {
		logServicesHaveDifferentNumberOfApplicationServers(serviceOne.IP, serviceOne.Port, serviceTwo.IP, serviceTwo.Port, len(serviceOne.ApplicationServers), len(serviceTwo.IP), modifyServiceName, id, modifyService.logging)
		return false
	}

	for _, applicationServerFromServiceOne := range serviceOne.ApplicationServers {
		var isFunded bool
		for _, applicationServerFromServiceTwo := range serviceTwo.ApplicationServers {
			if applicationServerFromServiceOne.IP == applicationServerFromServiceTwo.IP &&
				applicationServerFromServiceOne.Port == applicationServerFromServiceTwo.Port {
				isFunded = true
			}
		}
		if !isFunded {
			logApplicationServerNotFound(serviceOne.IP, serviceOne.Port, applicationServerFromServiceOne.IP, applicationServerFromServiceOne.Port, modifyServiceName, id, modifyService.logging)
			return false
		}
	}
	return true
}
