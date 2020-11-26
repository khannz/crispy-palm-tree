package healthcheck

import (
	"fmt"
	"strconv"
	"syscall"

	domain "github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

// PrepareDataForIPVS ...
func PrepareDataForIPVS(rawIP,
	rawPort,
	rawRoutingType,
	rawBalanceType,
	rawProtocol string,
	rawApplicationServers map[string]*domain.ApplicationServer) (string,
	uint16,
	uint32,
	string,
	uint16,
	map[string]uint16,
	error) {
	vip, port, routingType, balanceType, protocol, err := PrepareServiceForIPVS(rawIP,
		rawPort,
		rawRoutingType,
		rawBalanceType,
		rawProtocol)
	if err != nil {
		return "", 0, 0, "", 0, nil, err
	}

	applicationServers, err := convertRawApplicationServers(rawApplicationServers)
	if err != nil {
		return "", 0, 0, "", 0, nil, err
	}

	return vip, port, routingType, balanceType, protocol, applicationServers, nil
}

// PrepareServiceForIPVS ...
func PrepareServiceForIPVS(rawIP,
	rawPort,
	rawRoutingType,
	rawBalanceType,
	rawProtocol string) (string,
	uint16,
	uint32,
	string,
	uint16,
	error) {
	vip := rawIP
	port, err := stringToUINT16(rawPort)
	if err != nil {
		return "", 0, 0, "", 0, err
	}
	var routingType uint32
	switch rawRoutingType {
	case "masquerading":
		routingType = 0
	case "tunneling":
		routingType = 2
	default:
		return "", 0, 0, "", 0, fmt.Errorf("unknown routing type for prepare data for IPVS: %v", rawRoutingType)
	}
	balanceType := rawBalanceType
	protocol, err := protocolToUINT16(rawProtocol)
	if err != nil {
		return "", 0, 0, "", 0, err
	}

	return vip, port, routingType, balanceType, protocol, nil
}

func stringToUINT16(sval string) (uint16, error) {
	v, err := strconv.ParseUint(sval, 0, 16)
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

func protocolToUINT16(protocol string) (uint16, error) {
	switch protocol {
	case "tcp":
		return uint16(syscall.IPPROTO_TCP), nil
	case "udp":
		return uint16(syscall.IPPROTO_UDP), nil
	default:
		return uint16(0), fmt.Errorf("unknown protocol: %v", protocol)
	}
}

func convertRawApplicationServers(rawApplicationServers map[string]*domain.ApplicationServer) (map[string]uint16, error) {
	applicationServers := map[string]uint16{}

	for _, applicationServer := range rawApplicationServers {
		port, err := stringToUINT16(applicationServer.Port)
		if err != nil {
			return applicationServers, fmt.Errorf("can't convert port %v to type uint16: %v", applicationServer.Port, err)
		}
		applicationServers[applicationServer.IP] = port
	}
	return applicationServers, nil
}
