package application

// ServiceTransport ...
type ServiceTransport struct {
	IP              string `json:"ip"`
	Port            string `json:"port"`
	BalanceType     string `json:"balanceType"`
	RoutingType     string `json:"routingType"`
	Protocol        string `json:"protocol"`
	HealthcheckType string `json:"healthcheckType"`
	HelloTimer      string `json:"helloTimer"`
	ResponseTimer   string `json:"responseTimer"`
	AliveThreshold  string `json:"aliveThreshold"`
	DeadThreshold   string `json:"deadThreshold"`
	Quorum          string `json:"quorum"`
	// Hysteresis      string `json:"hysteresis"`
	ApplicationServersTransport []*ApplicationServerTransport `json:"-"`
}

// ApplicationServerTransport ...
type ApplicationServerTransport struct {
	IP                 string `json:"ip"`
	Port               string `json:"port"`
	HealthcheckAddress string `json:"healthcheckAddress"`
}
