package portadapter

import (
	"fmt"
	"sync"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

// MemoryWorker ...
type MemoryWorker struct {
	sync.Mutex

	Services                     domain.ServiceInfoConf
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

	return nil
}

func (memoryWorker *MemoryWorker) AddTunnelForApplicationServer(appSrvIP string) bool {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()
	_, isAppSrvIn := memoryWorker.ApplicationServersTunnelInfo[appSrvIP]
	if isAppSrvIn {
		memoryWorker.ApplicationServersTunnelInfo[appSrvIP]++
		return false
	}
	memoryWorker.ApplicationServersTunnelInfo[appSrvIP] = 1
	return true
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

func (memoryWorker *MemoryWorker) GetServices() domain.ServiceInfoConf {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()

	c := make(domain.ServiceInfoConf, len(memoryWorker.Services))
	for k, v := range memoryWorker.Services {
		c[k] = v
	}
	return c
}

func (memoryWorker *MemoryWorker) RemoveService(serviceInfo *domain.ServiceInfo) error {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()

	_, isServiceAlreadyIn := memoryWorker.Services[serviceInfo.Address]
	if !isServiceAlreadyIn {
		return fmt.Errorf("fail to remove service already don't have that %v", serviceInfo.Address)
	}

	delete(memoryWorker.Services, serviceInfo.Address)

	return nil
}

func (memoryWorker *MemoryWorker) RemoveTunnelForApplicationServer(appSrvIP string) bool {
	memoryWorker.Lock()
	defer memoryWorker.Unlock()

	if _, isAppSrvIn := memoryWorker.ApplicationServersTunnelInfo[appSrvIP]; !isAppSrvIn {
		return true
	}
	if memoryWorker.ApplicationServersTunnelInfo[appSrvIP] == 1 {
		delete(memoryWorker.Services, appSrvIP)
		return true
	}
	memoryWorker.ApplicationServersTunnelInfo[appSrvIP]--
	return false
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
