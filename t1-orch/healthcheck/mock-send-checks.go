package healthcheck

import (
	"time"
)

func IsTcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return false
}

func IsHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return false
}

func IsHttpAdvancedCheckOk(hcType string,
	healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return false
}

func IsIcmpCheckOk(ipS string,
	seq int,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return false
}
