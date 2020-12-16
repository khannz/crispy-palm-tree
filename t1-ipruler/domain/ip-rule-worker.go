package domain

// RouteWorker ...
type RouteWorker interface {
	AddIPRule(string, string) error
	RemoveIPRule(string, string) error
	GetIPRuleRuntimeConfig(string) (map[int]struct{}, error)
}
