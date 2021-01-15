package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type AddApplicationServersEntity struct {
	ipvs domain.IPVSWorker
}

func NewAddApplicationServersEntity(ipvs domain.IPVSWorker) *AddApplicationServersEntity {
	return &AddApplicationServersEntity{ipvs: ipvs}
}

func (addApplicationServersEntity *AddApplicationServersEntity) AddIPVSApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	return addApplicationServersEntity.ipvs.AddIPVSApplicationServersForService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id)
}
