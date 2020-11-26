package portadapter

type RouteEntity struct {
}

func NewRouteEntity() *RouteEntity {
	return &RouteEntity{}
}

func (routeEntity *RouteEntity) AddRoute(hcDestIP string, hcTunDestIP string, id string) error {
	return nil
}

func (routeEntity *RouteEntity) RemoveRoute(hcDestIP string, hcTunDestIP string, needRemoveTunnel bool, id string) error {
	return nil
}

func (routeEntity *RouteEntity) GetRouteRuntimeConfig(id string) ([]string, error) {
	return nil, nil
}
