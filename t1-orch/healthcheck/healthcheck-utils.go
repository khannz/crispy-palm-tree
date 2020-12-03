package healthcheck

import domain "github.com/khannz/crispy-palm-tree/t1-orch/domain"

func (hc *HeathcheckEntity) findServiceInHealtcheckSlice(address string) (int, bool) {
	var findedIndex int
	var isFinded bool
	for index, runningServiceHc := range hc.runningHeathchecks {
		if address == runningServiceHc.Address {
			findedIndex = index
			isFinded = true
			break
		}
	}
	return findedIndex, isFinded
}

func percentageOfUp(rawTotal, rawDown int) float32 {
	total := float32(rawTotal)
	down := float32(rawDown)
	if down == total {
		return 0
	}
	if down == 0 {
		return 100
	}
	return (total - down) * 100 / total
}

func percentageOfDownBelowMPercentOfAlivedForUp(pofUp float32, maxDownForUp int) bool {
	return float32(maxDownForUp) <= pofUp
}

func fillNewBooleanArray(newArray []bool, oldArray []bool) {
	if len(newArray) > len(oldArray) {
		reverceArrays(newArray, oldArray)
		for i := len(newArray) - 1; i >= 0; i-- {
			if i >= len(oldArray) {
				newArray[i] = false
			} else {
				newArray[i] = oldArray[i]
			}
		}
		reverceArrays(newArray, oldArray)
		return
	} else if len(newArray) == len(oldArray) {
		reverceArrays(newArray, oldArray)
		for i := range newArray {
			newArray[i] = oldArray[i]
		}
		reverceArrays(newArray, oldArray)
		return
	}

	reverceArrays(newArray, oldArray)
	tmpOldArraySlice := oldArray[:len(newArray)]
	copy(newArray, tmpOldArraySlice)
	reverceArrays(newArray, oldArray)
}

func reverceArrays(arOne, arTwo []bool) {
	reverceArray(arOne)
	reverceArray(arTwo)
}

func reverceArray(ar []bool) {
	for i, j := 0, len(ar)-1; i < j; i, j = i+1, j-1 {
		ar[i], ar[j] = ar[j], ar[i]
	}
}

func (hc *HeathcheckEntity) addNewServiceToMayAnnouncedServices(serviceIP string) {
	if _, inMap := hc.announcedServices[serviceIP]; inMap {
		hc.announcedServices[serviceIP]++
		return
	}
	hc.announcedServices[serviceIP] = 0 // add new service to annonced pool
}

func (hc *HeathcheckEntity) removeServiceFromMayAnnouncedServices(serviceIP string) {
	_, isFinded := hc.announcedServices[serviceIP]
	if isFinded {
		if hc.announcedServices[serviceIP] == 0 {
			delete(hc.announcedServices, serviceIP)
			return
		}
		hc.announcedServices[serviceIP]--
	}
}

func formDiffApplicationServers(newApplicationServers map[string]*domain.ApplicationServer, oldApplicationServers map[string]*domain.ApplicationServer) ([]string, []string, []string) {
	applicationServersForAdd := make([]string, 0)
	applicationServersForRemove := make([]string, 0)
	applicationServersAlreadyIN := make([]string, 0)
	for k := range newApplicationServers {
		if _, isFinded := oldApplicationServers[k]; isFinded {
			applicationServersAlreadyIN = append(applicationServersAlreadyIN, k)
			continue
		}
		applicationServersForAdd = append(applicationServersForAdd, k)
	}

	for k := range oldApplicationServers {
		if _, isFinded := newApplicationServers[k]; !isFinded {
			applicationServersForRemove = append(applicationServersForRemove, k)
		}
	}

	return applicationServersAlreadyIN, applicationServersForAdd, applicationServersForRemove
}

func getCopyOfApplicationServersFromService(serviceInfo *domain.ServiceInfo) map[string]*domain.ApplicationServer {
	applicationServers := make(map[string]*domain.ApplicationServer, len(serviceInfo.ApplicationServers))
	for k, applicationServer := range serviceInfo.ApplicationServers {
		tmpApplicationServerInternal := domain.InternalHC{
			HealthcheckType:    applicationServer.InternalHC.HealthcheckType,
			HealthcheckAddress: applicationServer.InternalHC.HealthcheckAddress,
			ResponseTimer:      applicationServer.InternalHC.ResponseTimer,
			AliveThreshold:     applicationServer.InternalHC.AliveThreshold,
			DeadThreshold:      applicationServer.InternalHC.DeadThreshold,
			LastIndexForAlive:  applicationServer.InternalHC.LastIndexForAlive,
			LastIndexForDead:   applicationServer.InternalHC.LastIndexForDead,
			NearFieldsMode:     applicationServer.InternalHC.NearFieldsMode,
			UserDefinedData:    applicationServer.InternalHC.UserDefinedData,
		}

		applicationServers[k] = &domain.ApplicationServer{
			Address:            applicationServer.Address,
			IP:                 applicationServer.IP,
			Port:               applicationServer.Port,
			IsUp:               applicationServer.IsUp,
			HealthcheckAddress: applicationServer.HealthcheckAddress,
			InternalHC:         tmpApplicationServerInternal,
		}
	}
	return applicationServers
}

func (hc *HeathcheckEntity) updateInStorage(hcService *domain.ServiceInfo, id string) {
	if err := hc.memoryWorker.UpdateService(hcService); err != nil {
		hc.logging.Warnf("can't update service %v from healtchecks: %v",
			hcService.Address,
			err)
	}
}

func (hc *HeathcheckEntity) moveApplicationServerStateIndexes(hcService *domain.ServiceInfo, applicationServerInfoKey string, isUpNow bool) {
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("moveApplicationServerStateIndexes: app srv index: %v,isUpNow: %v, total app srv at service: %v, AliveThreshold array len: %v, DeadThreshold array len: %v",
		applicationServerInfoKey,
		isUpNow,
		len(hcService.ApplicationServers),
		len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.AliveThreshold),
		len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.DeadThreshold))

	if len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.AliveThreshold) < hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForAlive+1 {
		hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForAlive = 0
	}
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.AliveThreshold[hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForAlive] = isUpNow
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForAlive++

	if len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.DeadThreshold) < hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDead+1 {
		hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDead = 0
	}
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.DeadThreshold[hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDead] = isUpNow
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDead++
}

func copyServiceInfo(hcService *domain.ServiceInfo) *domain.ServiceInfo {
	copyOfApplicationServersFromService := getCopyOfApplicationServersFromService(hcService)
	return &domain.ServiceInfo{
		Address:            hcService.Address,
		IP:                 hcService.IP,
		Port:               hcService.Port,
		IsUp:               hcService.IsUp,
		BalanceType:        hcService.BalanceType,
		RoutingType:        hcService.RoutingType,
		Protocol:           hcService.Protocol,
		Quorum:             hcService.Quorum,
		HealthcheckType:    hcService.HealthcheckType,
		HelloTimer:         hcService.HelloTimer,
		ResponseTimer:      hcService.ResponseTimer,
		HCNearFieldsMode:   hcService.HCNearFieldsMode,
		HCUserDefinedData:  hcService.HCUserDefinedData,
		AliveThreshold:     hcService.AliveThreshold,
		DeadThreshold:      hcService.DeadThreshold,
		ApplicationServers: copyOfApplicationServersFromService,
		HCStop:             make(chan struct{}, 1),
		HCStopped:          make(chan struct{}, 1),
	}
}
