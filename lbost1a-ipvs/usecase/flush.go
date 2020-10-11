package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type IPVSFlushEntity struct {
	ipvs domain.IPVSWorker
}

func NewIPVSFlushEntity(ipvs domain.IPVSWorker) *IPVSFlushEntity {
	return &IPVSFlushEntity{ipvs: ipvs}
}

func (iPVSFlushEntity *IPVSFlushEntity) IPVSFlush() error {
	return iPVSFlushEntity.ipvs.IPVSFlush()
}
