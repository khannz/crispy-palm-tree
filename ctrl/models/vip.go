package models

import "net"

type Vip struct {
	VipID          string
	SecurityZoneID int
	proto          string
	addr           net.IPAddr
	port           int
}
