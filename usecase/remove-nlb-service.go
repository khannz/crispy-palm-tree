package usecase

import (
	"fmt"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"github.com/sirupsen/logrus"
)

const removeNlbServiceEntity = "remove-nlb-service"

// RemoveNlbService ...
type RemoveNlbService struct {
	nwConfig         *domain.NetworkConfig
	tunnelConfig     domain.TunnelMaker
	keepalivedConfig domain.KeepalivedCustomizer
	uuidGenerator    domain.UUIDgenerator
	logging          *logrus.Logger
}

// NewRemoveNlbService ...
func NewRemoveNlbService(nwConfig *domain.NetworkConfig,
	tunnelConfig domain.TunnelMaker,
	keepalivedConfig domain.KeepalivedCustomizer,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveNlbService {
	return &RemoveNlbService{
		nwConfig:         nwConfig,
		tunnelConfig:     tunnelConfig,
		keepalivedConfig: keepalivedConfig,
		logging:          logging,
		uuidGenerator:    uuidGenerator,
	}
}

// RemoveNWBService ...
func (removeNlbService *RemoveNlbService) RemoveNWBService(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	removeRequestUUID string) error {
	removeNlbService.nwConfig.Lock()
	defer removeNlbService.nwConfig.Unlock()
	var err error
	deployedEntities, err := removeNlbService.findAlldeployedEntities(serviceIP, servicePort, applicationServers, removeRequestUUID)
	if err != nil {
		return fmt.Errorf("can't deployed entities: %v", err)
	}

	err = tunnelsRemove(deployedEntities, removeNlbService.tunnelConfig, removeRequestUUID)
	if err != nil {
		return fmt.Errorf("can't rollback tunnels: %v", err)
	}

	err = keepalivedConfigRemove(deployedEntities, removeNlbService.keepalivedConfig, removeRequestUUID)
	if err != nil {
		return fmt.Errorf("can't rollback keepalived configs: %v", err)
	}

	err = removeNlbService.keepalivedConfig.ReloadKeepalived(removeRequestUUID)
	if err != nil {
		return fmt.Errorf("can't reload keepalived: %v", err)
	}

	return nil
}

func (removeNlbService *RemoveNlbService) findAlldeployedEntities(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	removeRequestUUID string) (map[string][]string, error) {
	errors := []error{}
	deployedEntities := map[string][]string{}
	var err error

	deployedEntities, err = removeNlbService.tunnelConfig.DetectTunnels(applicationServers, deployedEntities, removeRequestUUID)
	if err != nil {
		errors = append(errors, err)
		err = nil
	}

	deployedEntities, err = removeNlbService.keepalivedConfig.DetectKeepalivedConfigFiles(serviceIP, servicePort, deployedEntities, removeRequestUUID)
	if err != nil {
		errors = append(errors, err)
		err = nil
	}

	deployedEntities, err = removeNlbService.keepalivedConfig.DetectKeepalivedConfigFileRows(serviceIP, servicePort, deployedEntities, removeRequestUUID)
	if err != nil {
		errors = append(errors, err)
		err = nil
	}

	return deployedEntities, combineErrors(errors)
}
