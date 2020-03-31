package application

import (
	"fmt"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/usecase"
	"github.com/sirupsen/logrus"
)

// BalancerFacade struct
type BalancerFacade struct {
	NetworkConfig        *domain.NetworkConfig
	TunnelConfig         domain.TunnelMaker
	KeepalivedCustomizer domain.KeepalivedCustomizer
	UUIDgenerator        domain.UUIDgenerator
	Logging              *logrus.Logger
}

// NewBalancerFacade ...
func NewBalancerFacade(networkConfig *domain.NetworkConfig,
	tunnelConfig domain.TunnelMaker,
	keepalivedCustomizer domain.KeepalivedCustomizer,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		NetworkConfig:        networkConfig,
		TunnelConfig:         tunnelConfig,
		KeepalivedCustomizer: keepalivedCustomizer,
		UUIDgenerator:        uuidGenerator,
		Logging:              logging,
	}
}

// NewNWBService ...
func (balancerFacade *BalancerFacade) NewNWBService(serviceIP, servicePort string, applicationServers map[string]string, newNWBRequestUUID string) error {
	addNlbService := usecase.NewAddNlbService(balancerFacade.NetworkConfig, balancerFacade.TunnelConfig, balancerFacade.KeepalivedCustomizer, balancerFacade.UUIDgenerator, balancerFacade.Logging)
	return addNlbService.CreateNewNWBService(serviceIP, servicePort, applicationServers, newNWBRequestUUID)
}

// RemoveNWBService ...
func (balancerFacade *BalancerFacade) RemoveNWBService(serviceIP, servicePort string, newNWBRequestUUID string) error {
	removeNWBService := usecase.NewRemoveNlbService(balancerFacade.NetworkConfig, balancerFacade.TunnelConfig, balancerFacade.KeepalivedCustomizer, balancerFacade.UUIDgenerator, balancerFacade.Logging)
	return removeNWBService.RemoveNWBService(serviceIP, servicePort, newNWBRequestUUID)
}

// GetNWBServices ...
func (balancerFacade *BalancerFacade) GetNWBServices(getNWBServicesUUID string) ([]domain.ServiceInfo, error) {
	getNWBServices := usecase.NewGetNlbServices(balancerFacade.NetworkConfig, balancerFacade.KeepalivedCustomizer, balancerFacade.Logging)
	nwbServices, err := getNWBServices.GetAllNWBServices(getNWBServicesUUID)
	if err != nil {
		return nil, fmt.Errorf("can't get nwb services: %v", err)
	}
	return nwbServices, nil
}

// AddApplicationServersToService ...
func (balancerFacade *BalancerFacade) AddApplicationServersToService(serviceIP, servicePort string, applicationServers map[string]string, newNWBRequestUUID string) (domain.ServiceInfo, error) {
	var err error
	var serviceInfo domain.ServiceInfo
	addApplicationServers := usecase.NewAddApplicationServers(balancerFacade.NetworkConfig, balancerFacade.TunnelConfig, balancerFacade.KeepalivedCustomizer, balancerFacade.UUIDgenerator, balancerFacade.Logging)
	serviceInfo, err = addApplicationServers.AddNewApplicationServers(serviceIP, servicePort, applicationServers, newNWBRequestUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't add application servers to service: %v", err)
	}
	return serviceInfo, nil
}
