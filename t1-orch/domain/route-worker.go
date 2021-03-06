package domain

// RouteWorker ...
type RouteWorker interface {
	AddRoute(string, string, string) error
	RemoveRoute(string, string, string) error
	GetRouteRuntimeConfig(string) ([]string, error)
}
