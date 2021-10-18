package domain

import (
	"math/big"
	"net"
)

//IPType ...
type IPType net.IP

//Int ...
func (ip IPType) Int() int64 {
	ipv4Int := big.NewInt(0)
	ipv4Int.SetBytes(net.IP(ip).To4())
	return ipv4Int.Int64()
}
