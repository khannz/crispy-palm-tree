package application

import (
	"time"

	domain "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/usecase"
	"github.com/sirupsen/logrus"
)

const sendRuntimeConfigName = "send runtime config"

// IPVSFacade struct
type IPVSFacade struct {
	IPVSWorker        domain.IPVSWorker
	HealthcheckWorker domain.HealthcheckWorker
	IDgenerator       domain.IDgenerator
	Logging           *logrus.Logger
}

// NewIPVSFacade ...
func NewIPVSFacade(ipvsWorker domain.IPVSWorker,
	healthcheckWorker domain.HealthcheckWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *IPVSFacade {

	return &IPVSFacade{
		IPVSWorker:        ipvsWorker,
		HealthcheckWorker: healthcheckWorker,
		IDgenerator:       idGenerator,
		Logging:           logging,
	}
}

func (ipvsFacade *IPVSFacade) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	id string) error {
	newNewIPVSServiceEntity := usecase.NewNewIPVSServiceEntity(ipvsFacade.IPVSWorker)
	return newNewIPVSServiceEntity.NewIPVSService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		id)
}

func (ipvsFacade *IPVSFacade) AddIPVSApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	newAddApplicationServersEntity := usecase.NewAddApplicationServersEntity(ipvsFacade.IPVSWorker)
	return newAddApplicationServersEntity.AddIPVSApplicationServersForService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id)
}

func (ipvsFacade *IPVSFacade) RemoveIPVSService(vip string,
	port uint16,
	protocol uint16,
	id string) error {
	newRemoveIPVSServiceEntity := usecase.NewRemoveIPVSServiceEntity(ipvsFacade.IPVSWorker)
	return newRemoveIPVSServiceEntity.RemoveIPVSService(vip,
		port,
		protocol,
		id)
}

func (ipvsFacade *IPVSFacade) RemoveIPVSApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	newRemoveIPVSApplicationServersFromServiceEntity := usecase.NewRemoveIPVSApplicationServersFromServiceEntity(ipvsFacade.IPVSWorker)
	return newRemoveIPVSApplicationServersFromServiceEntity.RemoveIPVSApplicationServersFromService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id)
}

func (ipvsFacade *IPVSFacade) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	newIsIPVSApplicationServerInService := usecase.NewGetIPVSRuntimeEntity(ipvsFacade.IPVSWorker)
	return newIsIPVSApplicationServerInService.GetIPVSRuntime(id)
}

func (ipvsFacade *IPVSFacade) TryToSendRuntimeConfig(id string) {
	newGetRuntimeConfigEntity := usecase.NewGetIPVSRuntimeEntity(ipvsFacade.IPVSWorker)
	newHealthcheckSenderEntity := usecase.NewHealthcheckSenderEntity(ipvsFacade.HealthcheckWorker)
	for {
		currentConfig, err := newGetRuntimeConfigEntity.GetIPVSRuntime(id)
		if err != nil {
			ipvsFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Errorf("failed to get runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := newHealthcheckSenderEntity.SendToHC(currentConfig, id); err != nil {
			ipvsFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Debugf("failed to send runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		ipvsFacade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": id,
		}).Info("send runtime config to hc success")
		break
	}
}
