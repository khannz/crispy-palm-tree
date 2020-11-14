package models

import "net"

type Real struct {
	RealID             string
	SecurityZoneID     int
	BalancingServiceID string
	addr               net.IPAddr
	port               int
}
