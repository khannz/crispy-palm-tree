package domain

import (
	"fmt"
	"strconv"
	"syscall"
)

// IPVSWorker ...
type IPVSWorker interface {
	CreateService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	RemoveService(string, uint16, uint16, string) error
	AddApplicationServersForService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	RemoveApplicationServersFromService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	Flush() error
}

// Utils

func PrepareDataForIPVS(rawIP,
	rawPort,
	rawRoutingType,
	rawBalanceType,
	rawProtocol string,
	rawApplicationServers []*ApplicationServer) (string,
	uint16,
	uint32,
	string,
	uint16,
	map[string]uint16,
	error) {
	vip := rawIP
	port, err := stringToUINT16(rawPort)
	if err != nil {
		return "", 0, 0, "", 0, nil, err
	}
	var routingType uint32
	switch rawRoutingType {
	case "masquerading":
		routingType = 0
	case "tunneling":
		routingType = 2
	default:
		return "", 0, 0, "", 0, nil, fmt.Errorf("unknown routing type for prepare data for IPVS: %v", routingType)
	}
	balanceType := rawBalanceType
	protocol, err := protocolToUINT16(rawProtocol)
	if err != nil {
		return "", 0, 0, "", 0, nil, err
	}
	applicationServers, err := convertRawApplicationServers(rawApplicationServers)
	if err != nil {
		return "", 0, 0, "", 0, nil, err
	}

	return vip, port, routingType, balanceType, protocol, applicationServers, nil
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

func convertRawApplicationServers(rawApplicationServers []*ApplicationServer) (map[string]uint16, error) {
	applicationServers := map[string]uint16{}

	for _, applicationServer := range rawApplicationServers {
		port, err := stringToUINT16(applicationServer.ServerPort)
		if err != nil {
			return applicationServers, fmt.Errorf("can't convert port %v to type uint16: %v", applicationServer.ServerPort, err)
		}
		applicationServers[applicationServer.ServerIP] = port
	}
	return applicationServers, nil
}
