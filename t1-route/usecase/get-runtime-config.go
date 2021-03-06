package usecase

import "github.com/khannz/crispy-palm-tree/t1-route/domain"

type GetRuntimeConfigEntity struct {
	routeWorker domain.RouteWorker
}

func NewGetRuntimeConfigEntity(routeWorker domain.RouteWorker) *GetRuntimeConfigEntity {
	return &GetRuntimeConfigEntity{routeWorker: routeWorker}
}

func (getRuntimeConfigEntity *GetRuntimeConfigEntity) GetRouteRuntimeConfig(id string) ([]string, error) {
	return getRuntimeConfigEntity.routeWorker.GetRouteRuntimeConfig(id)
}
