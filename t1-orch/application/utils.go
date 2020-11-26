package application

import (
	"math/big"
	"net"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

func enrichKVServiceDataToDomainServiceInfo(domainServiceInfo *domain.ServiceInfo) {
	hcAppSrvs := enrichKVApplicationServersDataToDomainApplicationServers(domainServiceInfo)
	domainServiceInfo.IsUp = false
	domainServiceInfo.ApplicationServers = hcAppSrvs
	domainServiceInfo.HCStop = make(chan struct{}, 1)
	domainServiceInfo.HCStopped = make(chan struct{}, 1)
}

func enrichKVApplicationServersDataToDomainApplicationServers(domainServiceInfo *domain.ServiceInfo) map[string]*domain.ApplicationServer {
	hass := make(map[string]*domain.ApplicationServer, len(domainServiceInfo.ApplicationServers))

	for addr, das := range domainServiceInfo.ApplicationServers {
		//
		ip, _, _ := net.ParseCIDR(das.IP)
		internalHC := domain.InternalHC{}
		internalHC.HCType = domainServiceInfo.HCType
		internalHC.HCAddress = domainServiceInfo.ApplicationServers[addr].HCAddress
		internalHC.HCTimeout = domainServiceInfo.HCTimeout
		internalHC.LastIndexForUp = 0
		internalHC.LastIndexForDown = 0
		internalHC.Mark = int(ip4toInt(ip))
		internalHC.NearFieldsMode = domainServiceInfo.HCNearFieldsMode
		internalHC.UserDefinedData = domainServiceInfo.HCUserDefinedData
		internalHC.RetriesForUP = make([]bool, domainServiceInfo.HCRetriesForUP)
		internalHC.RetriesForDown = make([]bool, domainServiceInfo.HCRetriesForDown)
		//
		has := &domain.ApplicationServer{
			Address:    addr,
			IP:         das.IP,
			Port:       das.Port,
			IsUp:       false,
			HCAddress:  das.HCAddress,
			InternalHC: internalHC,
		}
		hass[addr] = has
	}
	return hass
}

func ip4toInt(ipv4Address net.IP) int64 {
	ipv4Int := big.NewInt(0)
	ipv4Int.SetBytes(ipv4Address.To4())
	return ipv4Int.Int64()
}
