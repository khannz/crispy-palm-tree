package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type NewIPVSServiceEntity struct {
	ipvs domain.IPVSWorker
}

func NewNewIPVSServiceEntity(ipvs domain.IPVSWorker) *NewIPVSServiceEntity {
	return &NewIPVSServiceEntity{ipvs: ipvs}
}

func (NewIPVSServiceEntity *NewIPVSServiceEntity) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	id string) error {
	return NewIPVSServiceEntity.ipvs.NewIPVSService(
		id,
		balanceType,
		vip,
		protocol,
		port,
		routingType,
	)
}
