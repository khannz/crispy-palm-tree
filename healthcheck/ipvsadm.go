package healthcheck

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tehnerd/gnl2go"
	"github.com/thevan4/go-billet/executor"
)

// TODO:: possible bug, when ipvs app servers add/removed from healthcheckwhen that code running

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

// NewService ...
func (ipvsadmEntity *IPVSADMEntity) NewService(vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	createServiceID string) error {
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
	requestID string) error {
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
		ipvsadmEntity.logging.Debugf("at service %v:%v application server %v:%v removed from ipvs", serviceIP, servicePort, ip, port)
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
	updateServiceID string) error {
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
	updateServiceID string) error {
	ipvsadmEntity.Lock()
	defer ipvsadmEntity.Unlock()
	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	if err = ipvsadmEntity.removeApplicationServersFromService(ipvs, vip, port, protocol, applicationServers); err != nil {
		return fmt.Errorf("cant remove application servers from service: %v", err)
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
