package portadapter

import (
	"errors"
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
func (ipvsadmEntity *IPVSADMEntity) CreateService(serviceIP,
	rawServicePort string,
	rawApplicationServers map[string]string,
	createServiceUUID string) error {
	ipvsadmEntity.locker.Lock()
	defer ipvsadmEntity.locker.Unlock()

	ipvs, err := ipvsInit()
	if err != nil {
		return fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()

	err = checkServiceCreateIsValid(serviceIP,
		rawServicePort,
		rawApplicationServers,
		createServiceUUID)
	if err != nil {
		return fmt.Errorf("won't create service, got extra validate error: %v", err)
	}

	servicePort, err := stringToUINT16(rawServicePort)
	if err != nil {
		return fmt.Errorf("can't convert port stringToUINT16: %v", err)
	}

	applicationServers, err := convertRawApplicationServers(rawApplicationServers)
	if err != nil {
		return fmt.Errorf("can't convert application server port stringToUINT16: %v", err)
	}

	// AddService for IPv4
	err = ipvs.AddService(serviceIP, servicePort, uint16(gnl2go.ToProtoNum("tcp")), "rr")
	if err != nil {
		return fmt.Errorf("cant add ipv4 service AddService; err is : %v", err)
	}

	for ip, port := range applicationServers {
		err := ipvs.AddDestPort(serviceIP, servicePort, ip,
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
		return ipvs, fmt.Errorf("Cant initialize ipvs client, error is %v", err)
	}
	err = ipvs.Flush()
	if err != nil {
		return ipvs, fmt.Errorf("Error while running ipvs Flush method %v", err)
	}

	p, err := ipvs.GetPools()
	if err != nil {
		return ipvs, fmt.Errorf("Error while running ipvs GetPools method %v", err)
	}
	if len(p) != 0 {
		return ipvs, errors.New("ipvs Flush method havent cleared all the data")
	}
	return ipvs, nil
}

func checkServiceCreateIsValid(serviceIP,
	rawServicePort string,
	rawApplicationServers map[string]string,
	eventUUID string) error {
	var err error
	err = validateServiceUnique(serviceIP,
		rawServicePort,
		eventUUID)
	if err != nil {
		return fmt.Errorf("validateServiceUnique fail: %v", err)
	}

	err = validateApplicationServesIsUnique(rawApplicationServers, eventUUID)
	if err != nil {
		return fmt.Errorf("validateApplicationServesIsUnique fail: %v", err)
	}

	return nil
}

func validateServiceUnique(serviceIP,
	rawServicePort string,
	eventUUID string) error {
	//TODO: some validate
	return nil
}

func validateApplicationServesIsUnique(rawApplicationServers map[string]string,
	eventUUID string) error {
	var err error
	for ip, port := range rawApplicationServers {
		err = validateApplicationServerIsUnique(ip, port, eventUUID)
		if err != nil {
			return fmt.Errorf("validateApplicationServerIsUnique for ip %v port %v fail: %v", ip, port, err)
		}
	}
	return nil
}

func validateApplicationServerIsUnique(ip,
	port string,
	eventUUID string) error {
	//TODO: some validate
	return nil
}

func convertRawApplicationServers(rawApplicationServers map[string]string) (map[string]uint16, error) {
	applicationServers := map[string]uint16{}

	for ip, rawPort := range rawApplicationServers {
		port, err := stringToUINT16(rawPort)
		if err != nil {
			return applicationServers, fmt.Errorf("can't convert port %v to type uint16: %v", rawPort, err)
		}
		applicationServers[ip] = port
	}
	return applicationServers, nil
}
