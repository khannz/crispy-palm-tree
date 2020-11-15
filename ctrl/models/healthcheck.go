package models

type Healthcheck struct {
	HealthcheckID       string
	BalancingServiceID  string
	HealthcheckConfigID string
	HelloTimer          int
	ResponseTimer       int
	DeadThreshold       int
	AliveThreshold      int
	Quorum              int
	Hysteresis          int
}
