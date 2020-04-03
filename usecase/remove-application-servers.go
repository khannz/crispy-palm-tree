package usecase

import (
	"fmt"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"github.com/sirupsen/logrus"
)

const removeApplicationServers = "remove-application-servers"

// RemoveApplicationServers ...
type RemoveApplicationServers struct {
	nwConfig         *domain.NetworkConfig
	tunnelConfig     domain.TunnelMaker
	keepalivedConfig domain.KeepalivedCustomizer
	uuidGenerator    domain.UUIDgenerator
	logging          *logrus.Logger
}

// NewRemoveApplicationServers ...
func NewRemoveApplicationServers(nwConfig *domain.NetworkConfig,
	tunnelConfig domain.TunnelMaker,
	keepalivedConfig domain.KeepalivedCustomizer,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *RemoveApplicationServers {
	return &RemoveApplicationServers{
		nwConfig:         nwConfig,
		tunnelConfig:     tunnelConfig,
		keepalivedConfig: keepalivedConfig,
		logging:          logging,
		uuidGenerator:    uuidGenerator,
	}
}

// RemoveNewApplicationServers ...
func (removeApplicationServers *RemoveApplicationServers) RemoveNewApplicationServers(serviceIP,
	servicePort string,
	applicationServersForRemove map[string]string,
	removeApplicationServersUUID string) (domain.ServiceInfo, error) {
	removeApplicationServers.nwConfig.Lock()
	defer removeApplicationServers.nwConfig.Unlock()

	var serviceInfo domain.ServiceInfo

	currentKeepalivedConfigFile, currentApplicationServers, currentTunnelFilesForService, err := removeApplicationServers.getCurrentServiceConfig(serviceIP, servicePort, removeApplicationServersUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't get current service config: %v", err)
	}

	err = validateApplicationServersForRemove(applicationServersForRemove, currentApplicationServers)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't validate application servers for remove: %v", err)
	}
	err = removeApplicationServers.keepalivedConfig.RemoveApplicationServersFromKeepalivedConfigFile(currentKeepalivedConfigFile, applicationServersForRemove, removeApplicationServersUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't remove rows from keepalived config file: %v", err)
	}

	err = removeApplicationServers.tunnelConfig.RemoveApplicationServersFromTunnels(currentTunnelFilesForService, applicationServersForRemove, removeApplicationServersUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't remove application servers from tunnels: %v", err)
	}

	err = removeApplicationServers.keepalivedConfig.ReloadKeepalived(removeApplicationServersUUID)
	if err != nil {
		return serviceInfo, fmt.Errorf("can't reload keepalived: %v", err)
	}

	return serviceInfo, nil
}

func (removeApplicationServers *RemoveApplicationServers) getCurrentServiceConfig(serviceIP,
	servicePort string,
	requestUUID string) (string, []domain.ApplicationServer, []string, error) {

	currentKeepalivedConfigFile, err := removeApplicationServers.detectKeepalivedConfigWrapper(serviceIP, servicePort, requestUUID)
	if err != nil {
		return "", nil, nil, err
	}

	currentApplicationServers, err := removeApplicationServers.keepalivedConfig.GetApplicationServersByServiceIPAndPort(serviceIP, servicePort, requestUUID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("can't get application servers by service ip and port: %v", err)
	}

	currentTunnelFilesForService, err := removeApplicationServers.detectTunnelsWrapper(currentApplicationServers, requestUUID)
	if err != nil {
		return "", nil, nil, err
	}
	return currentKeepalivedConfigFile, currentApplicationServers, currentTunnelFilesForService, nil
}

func (removeApplicationServers *RemoveApplicationServers) detectTunnelsWrapper(applicationServers []domain.ApplicationServer, requestUUID string) ([]string, error) {
	deployedEntities := map[string][]string{}
	var err error
	deployedEntities, err = removeApplicationServers.tunnelConfig.DetectTunnels(applicationServers, deployedEntities, requestUUID)
	if err != nil {
		return nil, fmt.Errorf("can't detect tunnels: %v", err)
	}
	return deployedEntities["createdTunnelFiles"], nil
}

func (removeApplicationServers *RemoveApplicationServers) detectKeepalivedConfigWrapper(serviceIP, servicePort string, requestUUID string) (string, error) {
	deployedEntities := map[string][]string{}
	var err error
	deployedEntities, err = removeApplicationServers.keepalivedConfig.DetectKeepalivedConfigFiles(serviceIP, servicePort, deployedEntities, requestUUID)
	if err != nil {
		return "", fmt.Errorf("can't detectkeepalived config files: %v", err)
	}
	return deployedEntities["newFullKeepalivedDConfigFilePath"][0], nil
}
func validateApplicationServersForRemove(applicationServersForRemove map[string]string, currentApplicationServers []domain.ApplicationServer) error {
	if len(applicationServersForRemove) >= len(currentApplicationServers) {
		return fmt.Errorf("can't remove application servers. Current application servers %v, try remove %v application servers", len(currentApplicationServers), len(applicationServersForRemove))
	}
	return nil
}
