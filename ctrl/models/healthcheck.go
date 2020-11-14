package models

type Healthcheck struct {
	HealthcheckID       string
	BalancingServiceID  string
	HealthcheckConfigID string
	hello_timer         int
	response_timer      int
	dead_threshold      int
	alive_threshold     int
	quorum              int
	hysteresis          int
}
