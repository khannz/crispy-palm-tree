package usecase

import (
	"fmt"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"github.com/sirupsen/logrus"
)

const addNlbServiceEntity = "add-nlb-service"

// AddNlbService ...
type AddNlbService struct {
	nwConfig         *domain.NetworkConfig
	tunnelConfig     domain.TunnelMaker
	keepalivedConfig domain.KeepalivedCustomizer
	uuidGenerator    domain.UUIDgenerator
	logging          *logrus.Logger
}

// NewAddNlbService ...
func NewAddNlbService(nwConfig *domain.NetworkConfig,
	tunnelConfig domain.TunnelMaker,
	keepalivedConfig domain.KeepalivedCustomizer,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *AddNlbService {
	return &AddNlbService{
		nwConfig:         nwConfig,
		tunnelConfig:     tunnelConfig,
		keepalivedConfig: keepalivedConfig,
		logging:          logging,
		uuidGenerator:    uuidGenerator,
	}
}

// CreateNewNWBService ...
func (addNlbService *AddNlbService) CreateNewNWBService(serviceIP, servicePort string, realServersData map[string]string, newNWBRequestUUID string) error {
	addNlbService.nwConfig.Lock()
	defer addNlbService.nwConfig.Unlock()
	var err error
	deployedEntities := map[string][]string{}
	deployedEntities, err = addNlbService.tunnelConfig.CreateTunnel(deployedEntities, realServersData, newNWBRequestUUID)
	if err != nil {
		tunnelsRemove(deployedEntities, addNlbService.tunnelConfig, newNWBRequestUUID)
		return fmt.Errorf("Error when create tunnel: %v", err)
	}

	deployedEntities, err = addNlbService.keepalivedConfig.CustomizeKeepalived(serviceIP, servicePort, realServersData, deployedEntities, newNWBRequestUUID)
	if err != nil {
		tunnelsRemove(deployedEntities, addNlbService.tunnelConfig, newNWBRequestUUID)
		keepalivedConfigRemove(deployedEntities, addNlbService.keepalivedConfig, newNWBRequestUUID)
		addNlbService.keepalivedConfig.ReloadKeepalived(newNWBRequestUUID)
		return fmt.Errorf("Error when customize keepalived: %v", err)
	}
	return nil
}
