package application

import (
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
func (balancerFacade *BalancerFacade) NewNWBService(serviceIP, servicePort string, realServersData map[string]string, newNWBRequestUUID string) error {
	addNlbService := usecase.NewAddNlbService(balancerFacade.NetworkConfig, balancerFacade.TunnelConfig, balancerFacade.KeepalivedCustomizer, balancerFacade.UUIDgenerator, balancerFacade.Logging)
	return addNlbService.CreateNewNWBService(serviceIP, servicePort, realServersData, newNWBRequestUUID)
}

// RemoveNWBService ...
func (balancerFacade *BalancerFacade) RemoveNWBService(serviceIP, servicePort string, realServersData map[string]string, newNWBRequestUUID string) error {
	removeNWBService := usecase.NewRemoveNlbService(balancerFacade.NetworkConfig, balancerFacade.TunnelConfig, balancerFacade.KeepalivedCustomizer, balancerFacade.UUIDgenerator, balancerFacade.Logging)
	return removeNWBService.RemoveNWBService(serviceIP, servicePort, realServersData, newNWBRequestUUID)
}
