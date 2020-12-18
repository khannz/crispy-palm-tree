package usecase

import "github.com/khannz/crispy-palm-tree/t1-ipruler/domain"

type AddToRouteEntity struct {
	ipRuleWorker domain.IpRuleWorker
}

func NewAddToRouteEntity(ipRuleWorker domain.IpRuleWorker) *AddToRouteEntity {
	return &AddToRouteEntity{ipRuleWorker: ipRuleWorker}
}

func (addApplicationServersEntity *AddToRouteEntity) AddIPRule(hcTunDestIP string, id string) error {
	return addApplicationServersEntity.ipRuleWorker.AddIPRule(hcTunDestIP, id)
}
