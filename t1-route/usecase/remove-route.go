package usecase

import "github.com/khannz/crispy-palm-tree/t1-route/domain"

type RemoveRouteEntity struct {
	routeWorker domain.RouteWorker
}

func NewRemoveRouteEntity(routeWorker domain.RouteWorker) *RemoveRouteEntity {
	return &RemoveRouteEntity{routeWorker: routeWorker}
}

func (removeRouteEntity *RemoveRouteEntity) RemoveRoute(hcDestIP, hcTunDestIP string, id string) error {
	return removeRouteEntity.routeWorker.RemoveRoute(hcDestIP, hcTunDestIP, id)
}
