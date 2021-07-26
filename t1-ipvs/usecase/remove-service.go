package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type RemoveIPVSServiceEntity struct {
	ipvs domain.IPVSWorker
}

func NewRemoveIPVSServiceEntity(ipvs domain.IPVSWorker) *RemoveIPVSServiceEntity {
	return &RemoveIPVSServiceEntity{ipvs: ipvs}
}

func (removeIPVSServiceEntity *RemoveIPVSServiceEntity) RemoveIPVSService(vip string,
	port uint16,
	protocol uint16,
	id string) error {
	return removeIPVSServiceEntity.ipvs.RemoveIPVSService(
		id,
		vip,
		protocol,
		port,
	)
}
