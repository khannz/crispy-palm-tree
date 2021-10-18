package healthcheck

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

func percentageOfUp(rawTotal, rawDown int) float32 {
	if rawTotal == 0 || rawTotal < rawDown {
		return 0
	}
	return float32(rawTotal-rawDown) * 100. / float32(rawTotal)
}

func percentageOfDownBelowMPercentOfAlivedForUp(pofUp float32, maxDownForUp int) bool {
	return float32(maxDownForUp) <= pofUp
}

func fillNewBooleanArray(newArray []bool, oldArray []bool) {
	if len(newArray) > len(oldArray) {
		reverseArrays(newArray, oldArray)
		for i := len(newArray) - 1; i >= 0; i-- {
			if i >= len(oldArray) {
				newArray[i] = false
			} else {
				newArray[i] = oldArray[i]
			}
		}
		reverseArrays(newArray, oldArray)
		return
	} else if len(newArray) == len(oldArray) {
		reverseArrays(newArray, oldArray)
		for i := range newArray {
			newArray[i] = oldArray[i]
		}
		reverseArrays(newArray, oldArray)
		return
	}

	reverseArrays(newArray, oldArray)
	tmpOldArraySlice := oldArray[:len(newArray)]
	copy(newArray, tmpOldArraySlice)
	reverseArrays(newArray, oldArray)
}

func reverseArrays(arOne, arTwo []bool) {
	reverceArray(arOne)
	reverceArray(arTwo)
}

func reverceArray(ar []bool) {
	for i, j := 0, len(ar)-1; i < j; i, j = i+1, j-1 {
		ar[i], ar[j] = ar[j], ar[i]
	}
}

func formDiffApplicationServers(newApplicationServers domain.ApplicationServers, oldApplicationServers domain.ApplicationServers) ([]string, []string, []string) {
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
		if _, isFunded := newApplicationServers[k]; !isFunded {
			applicationServersForRemove = append(applicationServersForRemove, k)
		}
	}

	return applicationServersAlreadyIN, applicationServersForAdd, applicationServersForRemove
}

func getCopyOfApplicationServersFromService(serviceInfo *domain.ServiceInfo) domain.ApplicationServers {
	applicationServers := make(domain.ApplicationServers, len(serviceInfo.ApplicationServers))
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

func (hc *HealthcheckEntity) moveApplicationServerStateIndexes(hcService *domain.ServiceInfo, applicationServerInfoKey string, isUpNow bool) {
	hcService.Lock()
	defer hcService.Unlock()
	appServer := hcService.ApplicationServers[applicationServerInfoKey]
	if appServer == nil {
		return
	}
	internalHC := &appServer.InternalHC

	hc.logging.Tracef("moveApplicationServerStateIndexes: app srv index: %v,isUpNow: %v, total app srv at service: %v, AliveThreshold array len: %v, DeadThreshold array len: %v",
		applicationServerInfoKey,
		isUpNow,
		len(hcService.ApplicationServers),
		len(internalHC.AliveThreshold),
		len(internalHC.DeadThreshold))

	if len(internalHC.AliveThreshold) < internalHC.LastIndexForAlive+1 {
		internalHC.LastIndexForAlive = 0
	}
	internalHC.AliveThreshold[internalHC.LastIndexForAlive] = isUpNow
	internalHC.LastIndexForAlive++

	if len(internalHC.DeadThreshold) < internalHC.LastIndexForDead+1 {
		appServer.InternalHC.LastIndexForDead = 0
	}
	internalHC.DeadThreshold[internalHC.LastIndexForDead] = isUpNow
	internalHC.LastIndexForDead++
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

// Release stringer interface for print/log data in map[string]*dummyEntity
func (dummyEntity *dummyEntity) String() string {
	return fmt.Sprintf("Total for dummy: %v, announced for dummy: %v, is announced at dummy: %v",
		dummyEntity.totalForDummy,
		dummyEntity.announcedForDummy,
		dummyEntity.isAnnouncedAtDummy)
}
