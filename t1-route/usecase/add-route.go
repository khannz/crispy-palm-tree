package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-route/domain"

type AddToRouteEntity struct {
	routeWorker domain.RouteWorker
}

func NewAddToRouteEntity(routeWorker domain.RouteWorker) *AddToRouteEntity {
	return &AddToRouteEntity{routeWorker: routeWorker}
}

func (addApplicationServersEntity *AddToRouteEntity) AddRoute(hcDestIP, hcTunDestIP string, id string) error {
	return addApplicationServersEntity.routeWorker.AddRoute(hcDestIP, hcTunDestIP, id)
}
