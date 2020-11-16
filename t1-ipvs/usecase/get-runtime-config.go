package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type GetIPVSRuntimeEntity struct {
	ipvs domain.IPVSWorker
}

func NewGetIPVSRuntimeEntity(ipvs domain.IPVSWorker) *GetIPVSRuntimeEntity {
	return &GetIPVSRuntimeEntity{ipvs: ipvs}
}

func (GetIPVSRuntimeEntity *GetIPVSRuntimeEntity) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	return GetIPVSRuntimeEntity.ipvs.GetIPVSRuntime(id)
}
