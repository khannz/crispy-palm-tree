package portadapter

import (
	"fmt"
	"strconv"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/tehnerd/gnl2go"
	"github.com/thevan4/go-billet/executor"
)

// IPVSADMEntity ...
type IPVSADMEntity struct {
	locker *domain.Locker
}

// NewIPVSADMEntity ...
func NewIPVSADMEntity(locker *domain.Locker) (*IPVSADMEntity, error) {
	_, _, exitCode, err := executor.Execute("/usr/bin/ipvsadm", "", nil)
	if err != nil || exitCode != 0 {
		return nil, fmt.Errorf("got error when execute ipvsadm command: %v, exit code %v", err, exitCode)
	}
	return &IPVSADMEntity{locker: locker}, nil
}

// CreateService ... // TODO: also need protocol and balance type (weight?fwd IPVS_TUNNELING?)
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

	if err = ipvsadmEntity.addApplicationServersToService(ipvs, serviceInfo.ServiceIP, servicePort, applicationServers); err != nil {
		return fmt.Errorf("cant add application server to service: %v", err)
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

// ValidateHistoricalConfig ...
func (ipvsadmEntity *IPVSADMEntity) ValidateHistoricalConfig(storage *StorageEntity) error {

	pools, err := ipvsadmEntity.readActualConfig()
	if err != nil {
		return fmt.Errorf("can't read actual config: %v", err)
	}
	ipvsadmServicesInfo := transformRawIPVSPoolsToDomainModel(pools)
	storageServicesInfo, err := storage.LoadAllStorageDataToDomainModel()
	if err != nil {
		return fmt.Errorf("can't load all storage data to domain model: %v", err)
	}

	if err = compareDomainServicesData(ipvsadmServicesInfo, storageServicesInfo); err != nil {
		return fmt.Errorf("actual data does not match storage data: %v", err)
	}
	return nil
}

func (ipvsadmEntity *IPVSADMEntity) readActualConfig() ([]gnl2go.Pool, error) {
	ipvsadmEntity.locker.Lock()
	defer ipvsadmEntity.locker.Unlock()

	ipvs, err := ipvsInit()
	if err != nil {
		return nil, fmt.Errorf("can't ipvs Init: %v", err)
	}
	defer ipvs.Exit()
	pools, err := ipvs.GetPools()
	if err != nil {
		return nil, fmt.Errorf("ipvs can't get pools: %v", err)
	}
	return pools, nil
}

func transformRawIPVSPoolsToDomainModel(pools []gnl2go.Pool) []domain.ServiceInfo {
	servicesInfo := []domain.ServiceInfo{}
	for _, pool := range pools {
		applicationServers := []domain.ApplicationServer{}
		for _, dest := range pool.Dests {
			applocationServer := domain.ApplicationServer{
				ServerIP:   dest.IP,
				ServerPort: strconv.Itoa(int(dest.Port)),
			}
			applicationServers = append(applicationServers, applocationServer)
		}
		serviceInfo := domain.ServiceInfo{
			ServiceIP:          pool.Service.VIP,
			ServicePort:        strconv.Itoa(int(pool.Service.Port)),
			ApplicationServers: applicationServers,
		}
		servicesInfo = append(servicesInfo, serviceInfo)
	}
	return servicesInfo
}

func (ipvsadmEntity *IPVSADMEntity) addApplicationServersToService(ipvs *gnl2go.IpvsClient,
	serviceIP string, servicePort uint16,
	applicationServers map[string]uint16) error {
	for ip, port := range applicationServers {
		err := ipvs.AddDestPort(serviceIP, servicePort, ip,
			port, uint16(gnl2go.ToProtoNum("tcp")), 10, gnl2go.IPVS_TUNNELING)
		if err != nil {
			return fmt.Errorf("cant add dest to service sched flags: %v", err)
		}
	}
	return nil
}

func (ipvsadmEntity *IPVSADMEntity) removeApplicationServersFromService(ipvs *gnl2go.IpvsClient,
	serviceIP string, servicePort uint16,
	applicationServers map[string]uint16) error {
	for ip, port := range applicationServers {
		err := ipvs.DelDestPort(serviceIP, servicePort, ip,
			port, uint16(gnl2go.ToProtoNum("tcp")))
		if err != nil {
			return fmt.Errorf("cant add dest to service sched flags: %v", err)
		}
	}
	return nil
}

// AddApplicationServersFromService ...
func (ipvsadmEntity *IPVSADMEntity) AddApplicationServersFromService(serviceInfo domain.ServiceInfo,
	updateServiceUUID string) error {
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

	if err = ipvsadmEntity.addApplicationServersToService(ipvs, serviceInfo.ServiceIP, servicePort, applicationServers); err != nil {
		return fmt.Errorf("cant add application server to service: %v", err)
	}

	// TODO: log that ok
	return nil
}

// RemoveApplicationServersFromService ...
func (ipvsadmEntity *IPVSADMEntity) RemoveApplicationServersFromService(serviceInfo domain.ServiceInfo,
	updateServiceUUID string) error {
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

	if err = ipvsadmEntity.removeApplicationServersFromService(ipvs, serviceInfo.ServiceIP, servicePort, applicationServers); err != nil {
		return fmt.Errorf("cant add application server to service: %v", err)
	}

	// TODO: log that ok
	return nil
}
