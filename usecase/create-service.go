package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const createServiceName = "create-service"

// CreateServiceEntity ...
type CreateServiceEntity struct {
	locker           *domain.Locker
	configuratorVRRP domain.ServiceWorker
	tunnelConfig     domain.TunnelMaker
	uuidGenerator    domain.UUIDgenerator
	logging          *logrus.Logger
}

// NewCreateServiceEntity ...
func NewCreateServiceEntity(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	tunnelConfig domain.TunnelMaker,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *CreateServiceEntity {
	return &CreateServiceEntity{
		locker:           locker,
		configuratorVRRP: configuratorVRRP,
		tunnelConfig:     tunnelConfig,
		logging:          logging,
		uuidGenerator:    uuidGenerator,
	}
}

// CreateService ...
func (createService *CreateServiceEntity) CreateService(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	createServiceUUID string) error {
	// var err error
	// deployedEntities := map[string][]string{}
	// deployedEntities, err = createService.tunnelConfig.CreateTunnel(deployedEntities, applicationServers, newNWBRequestUUID)
	// if err != nil {
	// 	tunnelsRemove(deployedEntities, createService.tunnelConfig, newNWBRequestUUID)
	// 	return fmt.Errorf("Error when create tunnel: %v", err)
	// }
	err := createService.configuratorVRRP.CreateService(serviceIP, servicePort, applicationServers, createServiceUUID)
	if err != nil {
		return fmt.Errorf("Error when Configure VRRP: %v", err)
	}
	return nil
}
