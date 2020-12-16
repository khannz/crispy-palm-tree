package usecase

import "github.com/khannz/crispy-palm-tree/t1-ipruler/domain"

type RemoveIPRuleEntity struct {
	ipRuleWorker domain.RouteWorker
}

func NewRemoveIPRuleEntity(ipRuleWorker domain.RouteWorker) *RemoveIPRuleEntity {
	return &RemoveIPRuleEntity{ipRuleWorker: ipRuleWorker}
}

func (removeRouteEntity *RemoveIPRuleEntity) RemoveIPRule(hcTunDestIP string, id string) error {
	return removeRouteEntity.ipRuleWorker.RemoveIPRule(hcTunDestIP, id)
}
