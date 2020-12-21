package domain

// IpRuleWorker ...
type IpRuleWorker interface {
	AddIPRule(string, string) error
	RemoveIPRule(string, string) error
	GetIPRulerRuntime(string) (map[int]struct{}, error)
}
