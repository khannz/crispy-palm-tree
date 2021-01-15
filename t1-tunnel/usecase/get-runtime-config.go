package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"

type GetRuntimeConfigEntity struct {
	routeWorker domain.TunnelWorker
}

func NewGetRuntimeConfigEntity(routeWorker domain.TunnelWorker) *GetRuntimeConfigEntity {
	return &GetRuntimeConfigEntity{routeWorker: routeWorker}
}

func (getRuntimeConfigEntity *GetRuntimeConfigEntity) GetTunnelRuntime(id string) (map[string]struct{}, error) {
	return getRuntimeConfigEntity.routeWorker.GetTunnelRuntime(id)
}
