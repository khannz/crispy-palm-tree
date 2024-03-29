package domain

import "time"

// HealthcheckChecker generates a new id
type HealthcheckChecker interface {
	IsTcpCheckOk(string, time.Duration, int, string) bool
	IsHttpCheckOk(string, string, []int64, time.Duration, int, string) bool
	IsHttpsCheckOk(string, string, []int64, time.Duration, int, string) bool
	IsHttpAdvancedCheckOk(string, string, bool, map[string]string, time.Duration, int, string) bool
	IsIcmpCheckOk(string, time.Duration, int, string) bool
}
