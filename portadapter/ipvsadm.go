package portadapter

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/tehnerd/gnl2go"
)

// IPVSADMEntity ...
type IPVSADMEntity struct {
	locker *domain.Locker
}

// NewIPVSADMEntity ...
func NewIPVSADMEntity(locker *domain.Locker) *IPVSADMEntity {
	return &IPVSADMEntity{locker: locker}
}

// CreateService ... // FIXME: also need protocol and balance type (weight?fwd IPVS_TUNNELING?)
func (ipvsadmEntity *IPVSADMEntity) CreateService(serviceInfo domain.ServiceInfo,
	createServiceUUID string) error {
	ipvsadmEntity.locker.Lock()
	defer ipvsadmEntity.locker.Unlock()

	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	servicePort, err := stringToUINT16(serviceInfo.ServicePort)
	if err != nil {
		return fmt.Errorf("can't convert port stringToUINT16: %v", err)
	}

	applicationServers, err := convertRawApplicationServers(serviceInfo.ApplicationServers)
	if err != nil {
		return fmt.Errorf("can't convert application server port stringToUINT16: %v", err)
	}

	// AddService for IPv4
	err = ipvs.AddService(serviceInfo.ServiceIP, servicePort, uint16(gnl2go.ToProtoNum("tcp")), "rr")
	if err != nil {
		return fmt.Errorf("cant add ipv4 service AddService; err is : %v", err)
	}

	for ip, port := range applicationServers {
		err := ipvs.AddDestPort(serviceInfo.ServiceIP, servicePort, ip,
			port, uint16(gnl2go.ToProtoNum("tcp")), 10, gnl2go.IPVS_TUNNELING)
		if err != nil {
			return fmt.Errorf("cant add 1st dest to service sched flags: %v", err)
		}
	}
	// TODO: log that ok
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

func convertRawApplicationServers(rawApplicationServers []domain.ApplicationServer) (map[string]uint16, error) {
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

// RemoveService ...
func (ipvsadmEntity *IPVSADMEntity) RemoveService(serviceInfo domain.ServiceInfo, requestUUID string) error {
	ipvsadmEntity.locker.Lock()
	defer ipvsadmEntity.locker.Unlock()

	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	servicePort, err := stringToUINT16(serviceInfo.ServicePort)
	if err != nil {
		return fmt.Errorf("can't convert port stringToUINT16: %v", err)
	}

	err = ipvs.DelService(serviceInfo.ServiceIP, servicePort, uint16(gnl2go.ToProtoNum("tcp")))
	if err != nil {
		return fmt.Errorf("error while running DelService for ipv4: %v", err)
	}

	return nil
}
