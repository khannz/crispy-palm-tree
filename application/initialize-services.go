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
	for _, serviceConfigFromStorage := range servicesConfigsFromStorage {
		if err := balancerFacade.InitializeCreateService(serviceConfigFromStorage, uuid); err != nil {
			return err
		}
	}
	return nil
}

// InitializeCreateService ...
func (balancerFacade *BalancerFacade) InitializeCreateService(serviceConfigFromStorage *domain.ServiceInfo, uuid string) error {
	tunnelsFilesInfo := usecase.FormTunnelsFilesInfo(serviceConfigFromStorage.ApplicationServers, balancerFacade.CacheStorage)
	_, err := balancerFacade.TunnelConfig.CreateTunnels(tunnelsFilesInfo, uuid)
	if err != nil {
		return fmt.Errorf("can't create tunnel files: %v", err)
	}

	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(serviceConfigFromStorage.ServiceIP,
		serviceConfigFromStorage.ServicePort,
		serviceConfigFromStorage.RoutingType,
		serviceConfigFromStorage.BalanceType,
		serviceConfigFromStorage.Protocol,
		serviceConfigFromStorage.ApplicationServers)
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err := balancerFacade.IPVSADMConfigurator.CreateService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		uuid); err != nil {
		return fmt.Errorf("Error when ipvsadm create service: %v", err)
	}
	balancerFacade.HeathcheckEntity.NewServiceToHealtchecks(serviceConfigFromStorage)
	return nil
}
