package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"

type AddTunnelEntity struct {
	routeWorker domain.TunnelWorker
}

func NewAddTunnelEntity(routeWorker domain.TunnelWorker) *AddTunnelEntity {
	return &AddTunnelEntity{routeWorker: routeWorker}
}

func (addTunnelEntity *AddTunnelEntity) AddTunnel(hcTunDestIP string, id string) error {
	return addTunnelEntity.routeWorker.AddTunnel(hcTunDestIP, id)
}
