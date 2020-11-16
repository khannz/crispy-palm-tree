package portadapter

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tehnerd/gnl2go"
	"github.com/thevan4/go-billet/executor"
)

// FIXME: internal map of runtime, for exclude job errors

const (
	fallbackFlag           uint32 = 8
	fallbackAndshPortFlags uint32 = 24 // Bitwise OR fallbackFlag(8) | shPortFlag(16)
)

// IPVSADMEntity ...
type IPVSADMEntity struct {
	sync.Mutex
	logging *logrus.Logger
}

// NewIPVSADMEntity ...
func NewIPVSADMEntity(logging *logrus.Logger) (*IPVSADMEntity, error) {
	_, _, exitCode, err := executor.Execute("ipvsadm", "", nil)
	if err != nil || exitCode != 0 {
		return nil, fmt.Errorf("got error when execute ipvsadm command: %v, exit code %v", err, exitCode)
	}
	return &IPVSADMEntity{logging: logging}, nil
}

// NewIPVSService ...
func (ipvsadmEntity *IPVSADMEntity) NewIPVSService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	flags := chooseFlags(balanceType)
	// AddService for IPv4
	err = ipvs.AddServiceWithFlags(vip, port, protocol, balanceType, flags)
	if err != nil {
		return fmt.Errorf("cant add ipv4 service AddService; err is : %v", err)
	}

	if applicationServers != nil {
		if err = ipvsadmEntity.addApplicationServersToService(ipvs, vip, port, protocol, routingType, applicationServers); err != nil {
			return fmt.Errorf("cant add application server to service: %v", err)
		}
	}

	return nil
}

func ipvsInit() (*gnl2go.IpvsClient, error) {
	ipvs := new(gnl2go.IpvsClient)
	err := ipvs.Init()
	if err != nil {
		return ipvs, fmt.Errorf("cant initialize ipvs client, error is %v", err)
	}
	_, err = ipvs.GetPools()
	if err != nil {
		return ipvs, fmt.Errorf("error while running ipvs GetPools method %v", err)
	}

	return ipvs, nil
}

// RemoveIPVSService ...
func (ipvsadmEntity *IPVSADMEntity) RemoveIPVSService(vip string,
	port uint16,
	protocol uint16,
	id string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	errDel := ipvs.DelService(vip, port, protocol)
	if errDel != nil {
		return fmt.Errorf("error while running DelService for ipv4: %v", errDel)
	}

	return nil
}

func (ipvsadmEntity *IPVSADMEntity) addApplicationServersToService(ipvs *gnl2go.IpvsClient,
	serviceIP string, servicePort uint16, protocol uint16, routingType uint32,
	applicationServers map[string]uint16) error {
	for ip, port := range applicationServers {
		err := ipvs.AddDestPort(serviceIP, servicePort, ip,
			port, protocol, 10, routingType)
		if err != nil {
			return fmt.Errorf("cant add application server %v:%v to service %v:%v, protocol: %v, FWDMethod(routingType):%v",
				ip,
				port,
				serviceIP,
				servicePort,
				protocol,
				routingType)
		}
	}
	return nil
}

func (ipvsadmEntity *IPVSADMEntity) removeIPVSApplicationServersFromService(ipvs *gnl2go.IpvsClient,
	serviceIP string, servicePort uint16, protocol uint16,
	applicationServers map[string]uint16) error {
	for ip, port := range applicationServers {
		err := ipvs.DelDestPort(serviceIP, servicePort, ip,
			port, protocol)
		if err != nil {
			return fmt.Errorf("cant remove application server %v:%v from service %v:%v, protocol: %v",
				ip,
				port,
				serviceIP,
				servicePort,
				protocol)
		}
		ipvsadmEntity.logging.Debugf("at service %v:%v application server %v:%v removed from ipvs", serviceIP, servicePort, ip, port)
	}
	return nil
}

// AddIPVSApplicationServersForService ...
func (ipvsadmEntity *IPVSADMEntity) AddIPVSApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	if err = ipvsadmEntity.addApplicationServersToService(ipvs, vip, port, protocol, routingType, applicationServers); err != nil {
		return fmt.Errorf("cant add application server to service: %v", err)
	}

	return nil
}

// RemoveIPVSApplicationServersFromService ...
func (ipvsadmEntity *IPVSADMEntity) RemoveIPVSApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	if err = ipvsadmEntity.removeIPVSApplicationServersFromService(ipvs, vip, port, protocol, applicationServers); err != nil {
		return fmt.Errorf("cant remove application servers from service: %v", err)
	}

	return nil
}

// GetIPVSRuntime ...
func (ipvsadmEntity *IPVSADMEntity) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return nil, fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	pools, err := ipvs.GetPools()
	if err != nil {
		return nil, fmt.Errorf("can't read actual config: %v", err)
	}

	poolsMap := make(map[string]map[string]uint16, len(pools))
	for _, pool := range pools {
		applicationServers := make(map[string]uint16, len(pool.Dests))
		for _, dest := range pool.Dests {
			applicationServers[dest.IP] = dest.Port
		}
		poolsMap[pool.Service.VIP+":"+strconv.Itoa(int(pool.Service.Port))] = applicationServers
	}
	return poolsMap, nil
}

func chooseFlags(balanceType string) []byte {
	switch balanceType {
	case "mhf":
		return gnl2go.U32ToBinFlags(fallbackFlag)
	case "mhp":
		return gnl2go.U32ToBinFlags(fallbackAndshPortFlags)
	default:
		return nil
	}
}
