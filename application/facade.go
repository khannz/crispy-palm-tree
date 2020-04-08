package application

import (
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/usecase"
	"github.com/sirupsen/logrus"
)

// BalancerFacade struct
type BalancerFacade struct {
	Locker           *domain.Locker
	VRRPConfigurator domain.ServiceWorker
	TunnelConfig     domain.TunnelMaker
	UUIDgenerator    domain.UUIDgenerator
	Logging          *logrus.Logger
}

// NewBalancerFacade ...
func NewBalancerFacade(locker *domain.Locker,
	vrrpConfigurator domain.ServiceWorker,
	tunnelConfig domain.TunnelMaker,
	uuidGenerator domain.UUIDgenerator,
	logging *logrus.Logger) *BalancerFacade {

	return &BalancerFacade{
		Locker:           locker,
		VRRPConfigurator: vrrpConfigurator,
		TunnelConfig:     tunnelConfig,
		UUIDgenerator:    uuidGenerator,
		Logging:          logging,
	}
}

// CreateService ...
func (balancerFacade *BalancerFacade) CreateService(serviceIP, servicePort string, applicationServers map[string]string, newNWBRequestUUID string) error {
	newCreateServiceEntity := usecase.NewCreateServiceEntity(balancerFacade.Locker, balancerFacade.VRRPConfigurator, balancerFacade.TunnelConfig, balancerFacade.UUIDgenerator, balancerFacade.Logging)
	return newCreateServiceEntity.CreateService(serviceIP, servicePort, applicationServers, newNWBRequestUUID)
}

// // RemoveNWBService ...
// func (balancerFacade *BalancerFacade) RemoveNWBService(serviceIP, servicePort string, newNWBRequestUUID string) error {
// 	removeNWBService := usecase.NewRemoveNlbService(balancerFacade.TunnelConfig, balancerFacade.UUIDgenerator, balancerFacade.Logging)
// 	return removeNWBService.RemoveNWBService(serviceIP, servicePort, newNWBRequestUUID)
// }

// // GetNWBServices ...
// func (balancerFacade *BalancerFacade) GetNWBServices(getNWBServicesUUID string) ([]domain.ServiceInfo, error) {
// 	getNWBServices := usecase.NewGetNlbServices(balancerFacade.Logging)
// 	nwbServices, err := getNWBServices.GetAllNWBServices(getNWBServicesUUID)
// 	if err != nil {
// 		return nil, fmt.Errorf("can't get nwb services: %v", err)
// 	}
// 	return nwbServices, nil
// }

// // AddApplicationServersToService ...
// func (balancerFacade *BalancerFacade) AddApplicationServersToService(serviceIP, servicePort string, applicationServers map[string]string, newNWBRequestUUID string) (domain.ServiceInfo, error) {
// 	var err error
// 	var serviceInfo domain.ServiceInfo
// 	addApplicationServers := usecase.NewAddApplicationServers(balancerFacade.TunnelConfig, balancerFacade.UUIDgenerator, balancerFacade.Logging)
// 	serviceInfo, err = addApplicationServers.AddNewApplicationServers(serviceIP, servicePort, applicationServers, newNWBRequestUUID)
// 	if err != nil {
// 		return serviceInfo, fmt.Errorf("can't add application servers to service: %v", err)
// 	}
// 	return serviceInfo, nil
// }

// // RemoveApplicationServersFromService ...
// func (balancerFacade *BalancerFacade) RemoveApplicationServersFromService(serviceIP, servicePort string, applicationServers map[string]string, newNWBRequestUUID string) (domain.ServiceInfo, error) {
// 	var err error
// 	var serviceInfo domain.ServiceInfo
// 	removeApplicationServers := usecase.NewRemoveApplicationServers(balancerFacade.TunnelConfig, balancerFacade.UUIDgenerator, balancerFacade.Logging)
// 	serviceInfo, err = removeApplicationServers.RemoveNewApplicationServers(serviceIP, servicePort, applicationServers, newNWBRequestUUID)
// 	if err != nil {
// 		return serviceInfo, fmt.Errorf("can't remove application servers to service: %v", err)
// 	}
// 	return serviceInfo, nil
// }
