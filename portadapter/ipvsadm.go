package portadapter

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/tehnerd/gnl2go"
	"github.com/thevan4/go-billet/executor"
)

// TODO:: possible bug, when ipvs app servers add/removed from healthcheckwhen that code running

// IPVSADMEntity ...
type IPVSADMEntity struct {
	sync.Mutex
}

// NewIPVSADMEntity ...
func NewIPVSADMEntity() (*IPVSADMEntity, error) {
	_, _, exitCode, err := executor.Execute("ipvsadm", "", nil)
	if err != nil || exitCode != 0 {
		return nil, fmt.Errorf("got error when execute ipvsadm command: %v, exit code %v", err, exitCode)
	}
	return &IPVSADMEntity{}, nil
}

// CreateService ...
func (ipvsadmEntity *IPVSADMEntity) CreateService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	createServiceUUID string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	// AddService for IPv4
	err = ipvs.AddService(vip, port, protocol, balanceType)
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

// RemoveService ...
func (ipvsadmEntity *IPVSADMEntity) RemoveService(vip string,
	port uint16,
	protocol uint16,
	requestUUID string) error {
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

func transformRawIPVSPoolsToMap(pools []gnl2go.Pool) map[string]map[string]uint16 {
	servicesMapInfo := map[string]map[string]uint16{}
	for _, pool := range pools {
		applicationServers := map[string]uint16{}
		for _, dest := range pool.Dests {
			applicationServers[dest.IP] = dest.Port
		}
		servicesMapInfo[pool.Service.VIP+strconv.Itoa(int(pool.Service.Port))] = applicationServers
	}
	return servicesMapInfo
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

func (ipvsadmEntity *IPVSADMEntity) removeApplicationServersFromService(ipvs *gnl2go.IpvsClient,
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
	}
	return nil
}

// AddApplicationServersForService ...
func (ipvsadmEntity *IPVSADMEntity) AddApplicationServersForService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	updateServiceUUID string) error {
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

// RemoveApplicationServersFromService ...
func (ipvsadmEntity *IPVSADMEntity) RemoveApplicationServersFromService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	updateServiceUUID string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	pools, err := ipvs.GetPools()
	if err != nil {
		return fmt.Errorf("ipvs can't get pools: %v", err)
	}

	actualConfig := transformRawIPVSPoolsToMap(pools)
	if err != nil {
		return fmt.Errorf("can't read current config: %v", err)
	}

	vipAndPort := vip + strconv.Itoa(int(port))
	updatedApplicationServers := actualizesApplicationServersInCurrentConfig(actualConfig, vipAndPort, applicationServers)

	if updatedApplicationServers != nil {
		if err = ipvsadmEntity.removeApplicationServersFromService(ipvs, vip, port, protocol, updatedApplicationServers); err != nil {
			return fmt.Errorf("cant remove application server from service: %v", err) // TODO: do not return error?
		}
	}

	return nil
}

// Flush remove all ipvsadm data
func (ipvsadmEntity *IPVSADMEntity) Flush() error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	err = ipvs.Flush()
	if err != nil {
		return fmt.Errorf("can't ipvs Flush: %v", err)
	}
	return nil
}

// actualizesApplicationServersInCurrentConfig - actualizes application servers state
func actualizesApplicationServersInCurrentConfig(currentConfig map[string]map[string]uint16, incomeVipAndPort string, incomeApplicationServers map[string]uint16) map[string]uint16 {
	updatedApplicationServers := map[string]uint16{}
	if oldApplicationServers, serviceFinded := currentConfig[incomeVipAndPort]; serviceFinded {
		for incomeApplicationServerIP, incomeApplicationServerPort := range incomeApplicationServers {
			if oldApplicationServerPort, ipFinded := oldApplicationServers[incomeApplicationServerIP]; ipFinded {
				if oldApplicationServerPort == incomeApplicationServerPort {
					updatedApplicationServers[incomeApplicationServerIP] = incomeApplicationServerPort
				}
			}
		}
		return oldApplicationServers
	}
	return nil
}
