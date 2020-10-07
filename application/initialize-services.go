package application

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/healthcheck"
	"github.com/khannz/crispy-palm-tree/usecase"
)

// InitializeRuntimeSettings ...
func (balancerFacade *BalancerFacade) InitializeRuntimeSettings(id string) error {
	if err := balancerFacade.resetHealtchecksInfo(id); err != nil {
		return err
	}
	servicesConfigsFromStorage, err := balancerFacade.CacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return fmt.Errorf("fail to load  storage config at start")
	}
	if err := balancerFacade.removeOldTunnelsData(servicesConfigsFromStorage); err != nil {
		return fmt.Errorf("fail to remove old tunnels data: %v", err)
	}
	for _, serviceConfigFromStorage := range servicesConfigsFromStorage {
		if err := balancerFacade.InitializeCreateService(serviceConfigFromStorage, id); err != nil {
			return err
		}
	}
	return nil
}

func (balancerFacade *BalancerFacade) resetHealtchecksInfo(id string) error {
	servicesConfigsFromStorage, err := balancerFacade.CacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return fmt.Errorf("fail to load storage config at start: %v", err)
	}
	for i := range servicesConfigsFromStorage {
		servicesConfigsFromStorage[i].IsUp = false
		for j := range servicesConfigsFromStorage[i].ApplicationServers {
			servicesConfigsFromStorage[i].ApplicationServers[j].IsUp = false
		}
		if err := balancerFacade.CacheStorage.UpdateServiceInfo(servicesConfigsFromStorage[i], id); err != nil {
			return fmt.Errorf("fail to update service info at start")
		}
		if err := balancerFacade.PersistentStorage.UpdateServiceInfo(servicesConfigsFromStorage[i], id); err != nil {
			return fmt.Errorf("fail to update service info at start")
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
func (balancerFacade *BalancerFacade) InitializeCreateService(serviceConfigFromStorage *domain.ServiceInfo, id string) error {
	var tunnelsFilesInfo, newTunnelsFilesInfo []*domain.TunnelForApplicationServer
	var err error
	if serviceConfigFromStorage.Protocol == "tcp" { // TODO: too many if's, that dirty
		tunnelsFilesInfo = usecase.FormTunnelsFilesInfo(serviceConfigFromStorage.ApplicationServers, balancerFacade.CacheStorage)
		newTunnelsFilesInfo, err = balancerFacade.TunnelConfig.CreateTunnels(tunnelsFilesInfo, id)
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
	hcService := healthcheck.ConvertDomainServiceToHCService(serviceConfigFromStorage)
	balancerFacade.HeathcheckEntity.NewServiceToHealtchecks(hcService)
	return nil
}
