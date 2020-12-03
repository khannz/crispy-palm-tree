package portadapter

import (
	"fmt"
	"sync"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

// MemoryWorker ...
type MemoryWorker struct {
	sync.Mutex
	Services                     map[string]*domain.ServiceInfo
	ApplicationServersTunnelInfo map[string]int
}

func (memoryWorker *MemoryWorker) AddService(serviceInfo *domain.ServiceInfo) error {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()

	findedServiceInfo, isServiceAlreadyIn := memoryWorker.Services[serviceInfo.Address]
	if isServiceAlreadyIn {
		return fmt.Errorf("fail to add new service already have that %v: %v. need to use update service", serviceInfo.Address, findedServiceInfo)
	}

	memoryWorker.Services[serviceInfo.Address] = serviceInfo

	if serviceInfo.RoutingType == "tunneling" {
		for _, appSrv := range serviceInfo.ApplicationServers {
			memoryWorker.addApplicationServer(appSrv.IP)
		}
	}

	return nil
}

func (memoryWorker *MemoryWorker) addApplicationServer(appSrvIP string) {
	_, isAppSrvIn := memoryWorker.ApplicationServersTunnelInfo[appSrvIP]
	if isAppSrvIn {
		memoryWorker.ApplicationServersTunnelInfo[appSrvIP]++
		return
	}
	memoryWorker.ApplicationServersTunnelInfo[appSrvIP] = 1
}

func (memoryWorker *MemoryWorker) GetService(serviceAddress string) (*domain.ServiceInfo, error) {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()
	findedServiceInfo, isServiceIn := memoryWorker.Services[serviceAddress]
	if !isServiceIn {
		return nil, fmt.Errorf("service %v not found", serviceAddress)
	}
	return findedServiceInfo, nil
}

func (memoryWorker *MemoryWorker) GetServices() {
	// not implemented
	// memoryWorker.GetService("")
}

func (memoryWorker *MemoryWorker) RemoveService(serviceInfo *domain.ServiceInfo) error {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()

	_, isServiceAlreadyIn := memoryWorker.Services[serviceInfo.Address]
	if !isServiceAlreadyIn {
		return fmt.Errorf("fail to remove service already don't have that %v", serviceInfo.Address)
	}

	delete(memoryWorker.Services, serviceInfo.Address)

	if serviceInfo.RoutingType == "tunneling" {
		for _, appSrv := range serviceInfo.ApplicationServers {
			memoryWorker.removeApplicationServer(appSrv.IP)
		}
	}
	return nil
}

func (memoryWorker *MemoryWorker) removeApplicationServer(appSrvIP string) {
	if _, isAppSrvIn := memoryWorker.ApplicationServersTunnelInfo[appSrvIP]; !isAppSrvIn {
		return
	}
	if memoryWorker.ApplicationServersTunnelInfo[appSrvIP] == 1 {
		delete(memoryWorker.Services, appSrvIP)
		return
	}
	memoryWorker.ApplicationServersTunnelInfo[appSrvIP]--
}

func (memoryWorker *MemoryWorker) UpdateService(serviceInfo *domain.ServiceInfo) error {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()
	if _, isServiceIn := memoryWorker.Services[serviceInfo.Address]; !isServiceIn {
		return fmt.Errorf("service %v not found", serviceInfo.Address)
	}
	memoryWorker.Services[serviceInfo.Address] = serviceInfo
	return nil
}

func (memoryWorker *MemoryWorker) NeedTunnelForApplicationServer(appSrvIP string) bool {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()
	if _, isAppSrvIn := memoryWorker.ApplicationServersTunnelInfo[appSrvIP]; isAppSrvIn {
		return true
	}
	return false
}
