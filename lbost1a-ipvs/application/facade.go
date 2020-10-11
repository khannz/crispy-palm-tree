package application

import (
	domain "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/usecase"
	"github.com/sirupsen/logrus"
)

// IPVSFacade struct
type IPVSFacade struct {
	IPVSWorker  domain.IPVSWorker
	IDgenerator domain.IDgenerator
	Logging     *logrus.Logger
}

// NewIPVSFacade ...
func NewIPVSFacade(ipvsWorker domain.IPVSWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *IPVSFacade {

	return &IPVSFacade{
		IPVSWorker:  ipvsWorker,
		IDgenerator: idGenerator,
		Logging:     logging,
	}
}

func (ipvsFacade *IPVSFacade) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	newNewIPVSServiceEntity := usecase.NewNewIPVSServiceEntity(ipvsFacade.IPVSWorker)
	return newNewIPVSServiceEntity.NewIPVSService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
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

func (ipvsFacade *IPVSFacade) IsIPVSApplicationServerInService(serviceIP string,
	servicePort uint16,
	oneApplicationServerMap map[string]uint16,
	id string) (bool, error) {
	newIsIPVSApplicationServerInService := usecase.NewIsIPVSApplicationServerInService(ipvsFacade.IPVSWorker)
	return newIsIPVSApplicationServerInService.IsIPVSApplicationServerInService(serviceIP,
		servicePort,
		oneApplicationServerMap,
		id)
}
func (ipvsFacade *IPVSFacade) IPVSFlush() error {
	newIPVSFlushEntity := usecase.NewIPVSFlushEntity(ipvsFacade.IPVSWorker)
	return newIPVSFlushEntity.IPVSFlush()
}
