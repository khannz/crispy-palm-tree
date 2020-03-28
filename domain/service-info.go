package domain

// ServiceInfo ...
type ServiceInfo struct {
	ServiceIP       string
	ServicePort     string
	RealServers     []string
	HealthcheckType string
	ExtraInfo       map[string]interface{}
}
