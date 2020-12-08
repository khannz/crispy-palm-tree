package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"

type RemoveTunnelEntity struct {
	routeWorker domain.TunnelWorker
}

func NewRemoveTunnelEntity(routeWorker domain.TunnelWorker) *RemoveTunnelEntity {
	return &RemoveTunnelEntity{routeWorker: routeWorker}
}

func (removeTunnelEntity *RemoveTunnelEntity) RemoveTunnel(hcTunDestIP string, needRemoveTunnel bool, id string) error {
	return removeTunnelEntity.routeWorker.RemoveTunnel(hcTunDestIP, needRemoveTunnel, id)
}
