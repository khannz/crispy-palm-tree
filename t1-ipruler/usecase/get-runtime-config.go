package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipRule/domain"

type GetRuntimeConfigEntity struct {
	ipRuleWorker domain.RouteWorker
}

func NewGetRuntimeConfigEntity(ipRuleWorker domain.RouteWorker) *GetRuntimeConfigEntity {
	return &GetRuntimeConfigEntity{ipRuleWorker: ipRuleWorker}
}

func (getRuntimeConfigEntity *GetRuntimeConfigEntity) GetIPRuleRuntimeConfig(id string) (map[string]struct{}, error) {
	return getRuntimeConfigEntity.ipRuleWorker.GetIPRuleRuntimeConfig(id)
}
