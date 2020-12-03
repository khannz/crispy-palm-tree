package healthcheck

// TODO: hc to domain module
import (
	"fmt"
	"sync"

	domain "github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"

// HeathcheckEntity ...
type HeathcheckEntity struct {
	sync.Mutex
	runningHeathchecks []*domain.ServiceInfo // TODO: map much better
	memoryWorker       domain.MemoryWorker
	healthcheckChecker domain.HealthcheckChecker
	ipvsadm            domain.IPVSWorker
	dw                 domain.DummyWorker
	idGenerator        domain.IDgenerator
	logging            *logrus.Logger
	announcedServices  map[string]int
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(memoryWorker domain.MemoryWorker,
	healthcheckChecker domain.HealthcheckChecker,
	ipvsadm domain.IPVSWorker,
	dw domain.DummyWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HeathcheckEntity {

	return &HeathcheckEntity{
		runningHeathchecks: []*domain.ServiceInfo{}, // need for append
		memoryWorker:       memoryWorker,
		healthcheckChecker: healthcheckChecker,
		ipvsadm:            ipvsadm,
		dw:                 dw,
		idGenerator:        idGenerator,
		logging:            logging,
		announcedServices:  make(map[string]int),
	}
}

// NewServiceToHealtchecks - add service for healthchecks
func (hc *HeathcheckEntity) NewServiceToHealtchecks(hcService *domain.ServiceInfo, id string) error {
	hc.Lock()
	defer hc.Unlock()

	// hc.enrichApplicationServersHealthchecks(hcService, nil, false) // lock hcService
	hc.runningHeathchecks = append(hc.runningHeathchecks, hcService)
	hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
	if err := hc.addServiceToIPVS(hcService, id); err != nil {
		return fmt.Errorf("can't add srvice to IPVS: %v", err)
	}
	go hc.startHealthchecksForCurrentService(hcService)
	return nil
}

// RemoveServiceFromHealtchecks will work until removed
func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(hcService *domain.ServiceInfo, id string) error {
	hc.Lock()
	indexForRemove, isFinded := hc.findServiceInHealtcheckSlice(hcService.Address)
	hc.Unlock()
	if isFinded {
		hc.logging.Tracef("send stop checks for service %v", hcService.Address)
		hc.runningHeathchecks[indexForRemove].HCStop <- struct{}{}
		<-hc.runningHeathchecks[indexForRemove].HCStopped
		hc.annonceLogic(hcService.IP, false, id) // lock hc and dummy; may remove annonce
		hc.Lock()
		hc.removeServiceFromMayAnnouncedServices(hcService.IP) // hc must be locked
		hc.runningHeathchecks = append(hc.runningHeathchecks[:indexForRemove], hc.runningHeathchecks[indexForRemove+1:]...)
		hc.Unlock()
		if err := hc.removeServiceFromIPVS(hcService, id); err != nil {
			return fmt.Errorf("can't remove service from IPVS: %v", err)
		}
		hc.logging.Tracef("get checks stopped from service %v", hcService.Address)
		return nil
	}

	return fmt.Errorf("Heathcheck error: RemoveServiceFromHealtchecks error: service %v not found",
		hcService.Address)
}

// UpdateServiceAtHealtchecks ...
func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(hcService *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	hc.Lock()
	updateIndex, isFinded := hc.findServiceInHealtcheckSlice(hcService.Address) // find service for update at hc services
	hc.Unlock()
	if isFinded {
		hc.runningHeathchecks[updateIndex].Lock()
		hc.Lock()
		currentApplicationServers := getCopyOfApplicationServersFromService(hc.runningHeathchecks[updateIndex]) // copy of current services
		hc.Unlock()
		hc.runningHeathchecks[updateIndex].Unlock()
		hc.logging.Tracef("get update service job, sending stop checks for service %v", hcService.Address)
		hc.runningHeathchecks[updateIndex].HCStop <- struct{}{}
		<-hc.runningHeathchecks[updateIndex].HCStopped
		hc.logging.Tracef("healthchecks stopped for update service job %v", hcService.Address)
		hc.enrichApplicationServersHealthchecks(hcService, currentApplicationServers, hc.runningHeathchecks[updateIndex].IsUp) // lock hcService
		// do not include new app servers here! why??
		_, _, applicationServersForRemove := formDiffApplicationServers(hcService.ApplicationServers, currentApplicationServers)
		if len(applicationServersForRemove) != 0 {
			for _, k := range applicationServersForRemove {
				if err := hc.excludeApplicationServerFromIPVS(hcService, currentApplicationServers[k], id); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": id,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
		}

		tmpHC := append(hc.runningHeathchecks[:updateIndex], hcService)
		hc.Lock()
		hc.runningHeathchecks = append(tmpHC, hc.runningHeathchecks[updateIndex+1:]...)
		hc.Unlock()
		go hc.startHealthchecksForCurrentService(hcService)
		return hcService, nil
	}

	return hcService, fmt.Errorf("Heathcheck error: UpdateServiceAtHealtchecks error: service %v not found",
		hcService.Address)
}

func (hc *HeathcheckEntity) enrichApplicationServersHealthchecks(newServiceHealthcheck *domain.ServiceInfo, oldApplicationServers map[string]*domain.ApplicationServer, oldIsUpState bool) {
	newServiceHealthcheck.Lock()
	defer newServiceHealthcheck.Unlock()
	newServiceHealthcheck.IsUp = oldIsUpState
	newServiceHealthcheck.HCStop = make(chan struct{}, 1)
	newServiceHealthcheck.HCStopped = make(chan struct{}, 1)

	for k := range newServiceHealthcheck.ApplicationServers {
		newServiceHealthcheck.ApplicationServers[k].InternalHC.HealthcheckType = newServiceHealthcheck.HealthcheckType
		newServiceHealthcheck.ApplicationServers[k].InternalHC.HealthcheckAddress = newServiceHealthcheck.ApplicationServers[k].HealthcheckAddress
		newServiceHealthcheck.ApplicationServers[k].InternalHC.ResponseTimer = newServiceHealthcheck.ResponseTimer
		newServiceHealthcheck.ApplicationServers[k].InternalHC.LastIndexForAlive = 0
		newServiceHealthcheck.ApplicationServers[k].InternalHC.LastIndexForDead = 0
		newServiceHealthcheck.ApplicationServers[k].InternalHC.NearFieldsMode = newServiceHealthcheck.HCNearFieldsMode
		newServiceHealthcheck.ApplicationServers[k].InternalHC.UserDefinedData = newServiceHealthcheck.HCUserDefinedData

		retriesCounterForUp := make([]bool, newServiceHealthcheck.AliveThreshold)
		retriesCounterForDown := make([]bool, newServiceHealthcheck.DeadThreshold)

		if _, isFinded := oldApplicationServers[k]; isFinded {
			fillNewBooleanArray(retriesCounterForUp, oldApplicationServers[k].InternalHC.AliveThreshold)
			fillNewBooleanArray(retriesCounterForDown, oldApplicationServers[k].InternalHC.DeadThreshold)
			newServiceHealthcheck.ApplicationServers[k].InternalHC.AliveThreshold = retriesCounterForUp
			newServiceHealthcheck.ApplicationServers[k].InternalHC.DeadThreshold = retriesCounterForDown
			newServiceHealthcheck.ApplicationServers[k].IsUp = oldApplicationServers[k].IsUp
			hc.logging.Debugf("application server %v was found, is up state was moved", newServiceHealthcheck.ApplicationServers[k].Address)
			continue
		}
		newServiceHealthcheck.ApplicationServers[k].InternalHC.AliveThreshold = retriesCounterForUp
		newServiceHealthcheck.ApplicationServers[k].InternalHC.DeadThreshold = retriesCounterForDown
		newServiceHealthcheck.ApplicationServers[k].IsUp = false
		hc.logging.Debugf("application server %v NOT found, is up state set false", newServiceHealthcheck.ApplicationServers[k].Address)
	}
}

// annonceLogic - used when service change state

// // TODO: much faster if we have func isApplicationServersInIPSVSerrvice, return []int (indexes for app srv not fount)
// func (hc *HeathcheckEntity) isApplicationServerInIPSVService(hcServiceIP, rawHcServicePort, applicationServerIP, rawApplicationServerPort string, id string) bool {
// 	hcServicePort, err := stringToUINT16(rawHcServicePort)
// 	if err != nil {
// 		hc.logging.Errorf("can't convert port to uint16: %v", err)
// 	}
// 	applicationServerPort, err := stringToUINT16(rawApplicationServerPort)
// 	if err != nil {
// 		hc.logging.Errorf("can't convert port to uint16: %v", err)
// 	}

// 	oneAppSrvMap := make(map[string]uint16, 1)
// 	oneAppSrvMap[applicationServerIP] = applicationServerPort
// 	isApplicationServerInService, err := hc.ipvsadm.IsIPVSApplicationServerInService(hcServiceIP,
// 		hcServicePort,
// 		oneAppSrvMap,
// 		id)
// 	if err != nil {
// 		hc.logging.Errorf("can't check is application server in service: %v", err)
// 		return false
// 	}
// 	return isApplicationServerInService
// }

func (hc *HeathcheckEntity) GetServiceState(hcService *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	hc.Lock()
	defer hc.Unlock()
	for _, runHC := range hc.runningHeathchecks {
		if runHC.Address == hcService.Address {
			return copyServiceInfo(runHC), nil
		}
	}
	return nil, fmt.Errorf("get service state fail: service %v not found in healthchecks", hcService.Address)
}

func (hc *HeathcheckEntity) GetServicesState(id string) ([]*domain.ServiceInfo, error) {
	hc.Lock()
	defer hc.Unlock()
	if len(hc.runningHeathchecks) == 0 {
		return nil, fmt.Errorf("active hc services: %v", len(hc.runningHeathchecks))
	}
	copyOfServiceInfos := make([]*domain.ServiceInfo, len(hc.runningHeathchecks))
	for i, runHC := range hc.runningHeathchecks {
		copyOfServiceInfos[i] = copyServiceInfo(runHC)
	}
	return copyOfServiceInfos, nil
}
