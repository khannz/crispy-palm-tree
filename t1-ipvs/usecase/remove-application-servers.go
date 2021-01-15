package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type RemoveIPVSApplicationServersFromServiceEntity struct {
	ipvs domain.IPVSWorker
}

func NewRemoveIPVSApplicationServersFromServiceEntity(ipvs domain.IPVSWorker) *RemoveIPVSApplicationServersFromServiceEntity {
	return &RemoveIPVSApplicationServersFromServiceEntity{ipvs: ipvs}
}

func (removeIPVSApplicationServersFromServiceEntity *RemoveIPVSApplicationServersFromServiceEntity) RemoveIPVSApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	return removeIPVSApplicationServersFromServiceEntity.ipvs.RemoveIPVSApplicationServersFromService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id)
}
