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
	runningHeathchecks map[string]*domain.ServiceInfo // TODO: map much better
	// memoryWorker       domain.MemoryWorker
	healthcheckChecker domain.HealthcheckChecker
	ipvsadm            domain.IPVSWorker
	dw                 domain.DummyWorker
	idGenerator        domain.IDgenerator
	logging            *logrus.Logger
	announcedServices  map[string]int
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity( // memoryWorker domain.MemoryWorker,
	healthcheckChecker domain.HealthcheckChecker,
	ipvsadm domain.IPVSWorker,
	dw domain.DummyWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HeathcheckEntity {

	return &HeathcheckEntity{
		runningHeathchecks: map[string]*domain.ServiceInfo{}, // need for append
		// memoryWorker:       memoryWorker,
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

	hc.runningHeathchecks[hcService.Address] = hcService
	hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
	if err := hc.addServiceToIPVS(hcService, id); err != nil {
		return fmt.Errorf("can't add srvice to IPVS: %v", err)
	}
	go hc.startHealthchecksForCurrentService(hcService)
	return nil
}

// RemoveServiceFromHealtchecks will work until removed
func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(hcService *domain.ServiceInfo, id string) error {
	hc.Lock() // lock for find
	if _, isFinded := hc.runningHeathchecks[hcService.Address]; isFinded {
		hc.logging.Tracef("send stop checks for service %v", hcService.Address)
		hc.runningHeathchecks[hcService.Address].HCStop <- struct{}{}
		hc.Unlock() // unlock for find
		<-hc.runningHeathchecks[hcService.Address].HCStopped
		hc.annonceLogic(hcService.IP, false, id) // lock hc and dummy; may remove annonce
		hc.removeServiceFromMayAnnouncedServices(hcService.IP)
		hc.Lock() // lock for remove service
		delete(hc.runningHeathchecks, hcService.Address)
		hc.Unlock() // unlock for remove service
		if err := hc.removeServiceFromIPVS(hcService, id); err != nil {
			return fmt.Errorf("can't remove service from IPVS: %v", err)
		}
		hc.logging.Tracef("get checks stopped from service %v", hcService.Address)
		return nil
	}
	hc.Unlock() // unlock if service not found

	return fmt.Errorf("Heathcheck error: RemoveServiceFromHealtchecks error: service %v not found",
		hcService.Address)
}

// UpdateServiceAtHealtchecks ...
func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(hcService *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	hc.Lock()
	if _, isFinded := hc.runningHeathchecks[hcService.Address]; isFinded {
		currentApplicationServers := getCopyOfApplicationServersFromService(hc.runningHeathchecks[hcService.Address]) // copy of current services
		hc.logging.Tracef("get update service job, sending stop checks for service %v", hcService.Address)
		hc.runningHeathchecks[hcService.Address].HCStop <- struct{}{}
		hc.Unlock()
		<-hc.runningHeathchecks[hcService.Address].HCStopped
		hc.logging.Tracef("healthchecks stopped for update service job %v", hcService.Address)
		hc.enrichApplicationServersHealthchecks(hcService, currentApplicationServers, hc.runningHeathchecks[hcService.Address].IsUp) // lock hcService
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

		hc.Lock()
		hc.runningHeathchecks[hcService.Address] = hcService
		go hc.startHealthchecksForCurrentService(hcService)
		hc.Unlock()
		return hcService, nil
	}
	hc.Unlock() // unlock if not found

	return hcService, fmt.Errorf("Heathcheck error: UpdateServiceAtHealtchecks error: service %v not found",
		hcService.Address)
}

func (hc *HeathcheckEntity) enrichApplicationServersHealthchecks(newServiceHealthcheck *domain.ServiceInfo, oldApplicationServers map[string]*domain.ApplicationServer, oldIsUpState bool) {
	newServiceHealthcheck.Lock()
	defer newServiceHealthcheck.Unlock()
	newServiceHealthcheck.IsUp = oldIsUpState
	newServiceHealthcheck.HCStop = make(chan struct{}, 1)
	newServiceHealthcheck.HCStopped = make(chan struct{}, 1)
	// TODO: to many code for set InternalHC?
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
			hc.logging.Tracef("application server %v was found, is up state was moved", newServiceHealthcheck.ApplicationServers[k].Address)
			continue
		}
		newServiceHealthcheck.ApplicationServers[k].InternalHC.AliveThreshold = retriesCounterForUp
		newServiceHealthcheck.ApplicationServers[k].InternalHC.DeadThreshold = retriesCounterForDown
		newServiceHealthcheck.ApplicationServers[k].IsUp = false
		hc.logging.Tracef("application server %v NOT found, is up state set false", newServiceHealthcheck.ApplicationServers[k].Address)
	}
}

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

func (hc *HeathcheckEntity) GetServicesState(id string) (map[string]*domain.ServiceInfo, error) {
	hc.Lock()
	defer hc.Unlock()
	if len(hc.runningHeathchecks) == 0 {
		return nil, fmt.Errorf("active hc services: %v", len(hc.runningHeathchecks))
	}
	copyOfServiceInfos := make(map[string]*domain.ServiceInfo, len(hc.runningHeathchecks))
	for i, runHC := range hc.runningHeathchecks {
		copyOfServiceInfos[i] = copyServiceInfo(runHC)
	}
	return copyOfServiceInfos, nil
}
