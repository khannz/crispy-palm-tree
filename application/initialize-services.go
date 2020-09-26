package application

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/usecase"
)

// InitializeRuntimeSettings ...
func (balancerFacade *BalancerFacade) InitializeRuntimeSettings(uuid string) error {
	servicesConfigsFromStorage, err := balancerFacade.CacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return fmt.Errorf("fail to load  storage config at start")
	}
	if err := balancerFacade.removeOldTunnelsData(servicesConfigsFromStorage); err != nil {
		return fmt.Errorf("fail to remove old tunnels data: %v", err)
	}
	for _, serviceConfigFromStorage := range servicesConfigsFromStorage {
		if err := balancerFacade.InitializeCreateService(serviceConfigFromStorage, uuid); err != nil {
			return err
		}
	}
	return nil
}

func (balancerFacade *BalancerFacade) removeOldTunnelsData(servicesConfigsFromStorage []*domain.ServiceInfo) error {
	allApplicationServersIPWhoHaveTunnels, err := balancerFacade.CacheStorage.GetAllApplicationServersIPWhoHaveTunnels()
	if err != nil {
		return fmt.Errorf("can't get all application servers ip who have tunnels:%v", err)
	}
	for _, allApplicationServerIPWhoHaveTunnels := range allApplicationServersIPWhoHaveTunnels {
		if err := balancerFacade.CacheStorage.RemoveTunnelsInfoForApplicationServerFromStorage(allApplicationServerIPWhoHaveTunnels); err != nil {
			return fmt.Errorf("can't remove tunnel info from cache storage for application server %v, got error: %v", allApplicationServerIPWhoHaveTunnels, err)
		}
		if err := balancerFacade.PersistentStorage.RemoveTunnelsInfoForApplicationServerFromStorage(allApplicationServerIPWhoHaveTunnels); err != nil {
			return fmt.Errorf("can't remove tunnel info from persistent storage for application server %v, got error: %v", allApplicationServerIPWhoHaveTunnels, err)
		}
	}
	return nil
}

// InitializeCreateService ...
func (balancerFacade *BalancerFacade) InitializeCreateService(serviceConfigFromStorage *domain.ServiceInfo, uuid string) error {
	var tunnelsFilesInfo, newTunnelsFilesInfo []*domain.TunnelForApplicationServer
	var err error
	if serviceConfigFromStorage.Protocol == "tcp" { // TODO: too many if's, that dirty
		tunnelsFilesInfo = usecase.FormTunnelsFilesInfo(serviceConfigFromStorage.ApplicationServers, balancerFacade.CacheStorage)
		newTunnelsFilesInfo, err = balancerFacade.TunnelConfig.CreateTunnels(tunnelsFilesInfo, uuid)
		if err != nil {
			return fmt.Errorf("can't create tunnel files: %v", err)
		}
		balancerFacade.Logging.Infof("new tunnels created: %v", tunnelsFilesInfo)

		if err := balancerFacade.CacheStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return fmt.Errorf("can't add to cache storage :%v", err)
		}
		if err := balancerFacade.PersistentStorage.UpdateTunnelFilesInfoAtStorage(newTunnelsFilesInfo); err != nil {
			return fmt.Errorf("can't add to persistent storage :%v", err)
		}
	}
	vip, port, routingType, balanceType, protocol, err := domain.PrepareServiceForIPVS(serviceConfigFromStorage.ServiceIP,
		serviceConfigFromStorage.ServicePort,
		serviceConfigFromStorage.RoutingType,
		serviceConfigFromStorage.BalanceType,
		serviceConfigFromStorage.Protocol)
	if err != nil {
		return fmt.Errorf("error prepare data for IPVS: %v", err)
	}
	if err := balancerFacade.IPVSADMConfigurator.CreateService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		nil,
		uuid); err != nil {
		return fmt.Errorf("Error when ipvsadm create service: %v", err)
	}
	balancerFacade.HeathcheckEntity.NewServiceToHealtchecks(serviceConfigFromStorage)
	return nil
}
