package usecase

// import (
// 	"fmt"

// 	"github.com/khannz/crispy-palm-tree/domain"
// 	"github.com/sirupsen/logrus"
// )

// const addApplicationServers = "add-application-servers"

// // AddApplicationServers ...
// type AddApplicationServers struct {
// 	nwConfig         *domain.NetworkConfig
// 	tunnelConfig     domain.TunnelMaker
// 	keepalivedConfig domain.KeepalivedCustomizer
// 	uuidGenerator    domain.UUIDgenerator
// 	logging          *logrus.Logger
// }

// // NewAddApplicationServers ...
// func NewAddApplicationServers(nwConfig *domain.NetworkConfig,
// 	tunnelConfig domain.TunnelMaker,
// 	keepalivedConfig domain.KeepalivedCustomizer,
// 	uuidGenerator domain.UUIDgenerator,
// 	logging *logrus.Logger) *AddApplicationServers {
// 	return &AddApplicationServers{
// 		nwConfig:         nwConfig,
// 		tunnelConfig:     tunnelConfig,
// 		keepalivedConfig: keepalivedConfig,
// 		logging:          logging,
// 		uuidGenerator:    uuidGenerator,
// 	}
// }

// // AddNewApplicationServers ...
// func (addApplicationServers *AddApplicationServers) AddNewApplicationServers(serviceIP,
// 	servicePort string,
// 	applicationServers map[string]string,
// 	addApplicationServersUUID string) (domain.ServiceInfo, error) {
// 	addApplicationServers.nwConfig.Lock()
// 	defer addApplicationServers.nwConfig.Unlock()
// 	var err error
// 	var serviceInfo domain.ServiceInfo
// 	deployedEntities := map[string][]string{}
// 	deployedEntities, err = addApplicationServers.tunnelConfig.CreateTunnel(deployedEntities, applicationServers, addApplicationServersUUID)
// 	if err != nil {
// 		tunnelsRemove(deployedEntities, addApplicationServers.tunnelConfig, addApplicationServersUUID)
// 		return serviceInfo, fmt.Errorf("Error when create tunnel: %v", err)
// 	}

// 	serviceInfo, deployedEntities, err = addApplicationServers.keepalivedConfig.AddApplicationServersToKeepalived(serviceIP, servicePort, applicationServers, deployedEntities, addApplicationServersUUID)
// 	if err != nil {
// 		tunnelsRemove(deployedEntities, addApplicationServers.tunnelConfig, addApplicationServersUUID)
// 		// keepalivedConfigRemoveRows(deployedEntities, addApplicationServers.keepalivedConfig, addApplicationServersUUID)
// 		addApplicationServers.keepalivedConfig.ReloadKeepalived(addApplicationServersUUID)
// 		return serviceInfo, fmt.Errorf("Error when customize keepalived: %v", err)
// 	}
// 	return serviceInfo, nil
// }
