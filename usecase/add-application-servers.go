package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const addApplicationServersName = "add-application-servers"

// AddApplicationServers ...
type AddApplicationServers struct {
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

// NewAddApplicationServers ...
func NewAddApplicationServers(locker *domain.Locker,
	ipvsadm domain.IPVSWorker,
	cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	tunnelConfig domain.TunnelMaker,
	hc domain.HeathcheckWorker,
	commandGenerator domain.CommandGenerator,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *AddApplicationServers {
	return &AddApplicationServers{
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

// AddNewApplicationServers ...
func (addApplicationServers *AddApplicationServers) AddNewApplicationServers(newServiceInfo *domain.ServiceInfo,
	addApplicationServersID string) (*domain.ServiceInfo, error) {
	var err error
	var updatedServiceInfo *domain.ServiceInfo

	// graceful shutdown part start
	addApplicationServers.locker.Lock()
	defer addApplicationServers.locker.Unlock()
	addApplicationServers.gracefulShutdown.Lock()
	if addApplicationServers.gracefulShutdown.ShutdownNow {
		defer addApplicationServers.gracefulShutdown.Unlock()
		return newServiceInfo, fmt.Errorf("program got shutdown signal, job add application servers %v cancel", newServiceInfo)
	}
	addApplicationServers.gracefulShutdown.UsecasesJobs++
	addApplicationServers.gracefulShutdown.Unlock()
	defer decreaseJobs(addApplicationServers.gracefulShutdown)
	// graceful shutdown part end

	logStartUsecase(addApplicationServersName, "add new application servers to service", addApplicationServersID, newServiceInfo, addApplicationServers.logging)
	logTryPreValidateRequest(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	allCurrentServices, err := addApplicationServers.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return newServiceInfo, fmt.Errorf("fail when loading info about current services: %v", err)
	}
	if !isServiceExist(newServiceInfo.ServiceIP, newServiceInfo.ServicePort, allCurrentServices) {
		return newServiceInfo, fmt.Errorf("service %v:%v does not exist, can't add application servers", newServiceInfo.ServiceIP, newServiceInfo.ServicePort)
	}

	if err = checkApplicationServersIPAndPortUnique(newServiceInfo.ApplicationServers, allCurrentServices); err != nil {
		return newServiceInfo, err
	}

	logTryToGetCurrentServiceInfo(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	currentServiceInfo, err := addApplicationServers.cacheStorage.GetServiceInfo(newServiceInfo, addApplicationServersID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't get service info: %v", err)
	}
	newServiceInfo.RoutingType = currentServiceInfo.RoutingType // for ipvs and for check routing type is valid
	newServiceInfo.Protocol = currentServiceInfo.Protocol
	logGotCurrentServiceInfo(addApplicationServersName, addApplicationServersID, currentServiceInfo, addApplicationServers.logging)

	if err = checkRoutingTypeForApplicationServersValid(newServiceInfo, allCurrentServices); err != nil {
		return newServiceInfo, err
	}
	logPreValidateRequestIsOk(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)

	var tunnelsFilesInfo, newTunnelsFilesInfo []*domain.TunnelForApplicationServer
	if currentServiceInfo.Protocol == "tcp" {
		tunnelsFilesInfo = FormTunnelsFilesInfo(newServiceInfo.ApplicationServers, addApplicationServers.cacheStorage)
		logTryCreateNewTunnels(addApplicationServersName, addApplicationServersID, tunnelsFilesInfo, addApplicationServers.logging)
		newTunnelsFilesInfo, err = addApplicationServers.tunnelConfig.CreateTunnels(tunnelsFilesInfo, addApplicationServersID)
		if err != nil {
			return nil, fmt.Errorf("can't create tunnel files: %v", err)
		}
		logCreatedNewTunnels(addApplicationServersName, addApplicationServersID, tunnelsFilesInfo, addApplicationServers.logging)

		if err := addApplicationServers.cacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return updatedServiceInfo, fmt.Errorf("can't update tunnel info")
		}
	}

	logTryGenerateUpdatedServiceInfo(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	updatedServiceInfo, err = forAddApplicationServersFormUpdateServiceInfo(currentServiceInfo, newServiceInfo, addApplicationServersID)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("can't form update service info: %v", err)
	}
	logGenerateUpdatedServiceInfo(addApplicationServersName, addApplicationServersID, updatedServiceInfo, addApplicationServers.logging)

	logTryIpvsadmApplicationServers(addApplicationServersName, addApplicationServersID, newServiceInfo.ApplicationServers, newServiceInfo.ServiceIP, newServiceInfo.ServicePort, addApplicationServers.logging)
	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(currentServiceInfo.ServiceIP,
		currentServiceInfo.ServicePort,
		currentServiceInfo.RoutingType,
		currentServiceInfo.BalanceType,
		currentServiceInfo.Protocol,
		newServiceInfo.ApplicationServers)
	if err != nil {
		return updatedServiceInfo, fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err = addApplicationServers.ipvsadm.AddApplicationServersForService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		addApplicationServersID); err != nil {
		return currentServiceInfo, fmt.Errorf("Error when ipvsadm add application servers for service: %v", err)
	}
	logAddedIpvsadmApplicationServers(addApplicationServersName, addApplicationServersID, newServiceInfo.ApplicationServers, newServiceInfo.ServiceIP, newServiceInfo.ServicePort, addApplicationServers.logging)

	// add to cache storage
	logTryUpdateServiceInfoAtCache(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	if err = addApplicationServers.cacheStorage.UpdateServiceInfo(updatedServiceInfo, addApplicationServersID); err != nil {
		return currentServiceInfo, fmt.Errorf("can't add to cache storage: %v", err)
	}
	logUpdateServiceInfoAtCache(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)

	logTryUpdateServiceInfoAtPersistentStorage(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	if err = addApplicationServers.persistentStorage.UpdateServiceInfo(updatedServiceInfo, addApplicationServersID); err != nil {
		return currentServiceInfo, fmt.Errorf("error when update persistent storage: %v", err)
	}
	logUpdatedServiceInfoAtPersistentStorage(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	if currentServiceInfo.Protocol == "tcp" {
		if err := addApplicationServers.persistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return updatedServiceInfo, fmt.Errorf("can't update tunnel info")
		}
	}
	logTryGenerateCommandsForApplicationServers(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	if err := addApplicationServers.commandGenerator.GenerateCommandsForApplicationServers(updatedServiceInfo, addApplicationServersID); err != nil {
		return updatedServiceInfo, fmt.Errorf("can't generate commands :%v", err)
	}
	logGeneratedCommandsForApplicationServers(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)

	logUpdateServiceAtHealtchecks(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	if err = addApplicationServers.hc.UpdateServiceAtHealtchecks(updatedServiceInfo); err != nil {
		return updatedServiceInfo, fmt.Errorf("application server added, but not activated, an error occurred when adding to the healtchecks: %v", err)
	}
	logUpdatedServiceAtHealtchecks(addApplicationServersName, addApplicationServersID, addApplicationServers.logging)
	return updatedServiceInfo, nil
}
