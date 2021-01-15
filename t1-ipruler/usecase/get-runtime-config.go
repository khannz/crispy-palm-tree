package usecase

import "github.com/khannz/crispy-palm-tree/t1-ipruler/domain"

type GetRuntimeConfigEntity struct {
	ipRuleWorker domain.IpRuleWorker
}

func NewGetRuntimeConfigEntity(ipRuleWorker domain.IpRuleWorker) *GetRuntimeConfigEntity {
	return &GetRuntimeConfigEntity{ipRuleWorker: ipRuleWorker}
}

func (getRuntimeConfigEntity *GetRuntimeConfigEntity) GetIPRuleRuntimeConfig(id string) (map[int]struct{}, error) {
	return getRuntimeConfigEntity.ipRuleWorker.GetIPRuleRuntimeConfig(id)
}
