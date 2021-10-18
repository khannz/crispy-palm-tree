package healthcheck

// TODO: hc to domain module
import (
	"fmt"
	"sync"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"

// HealthcheckEntity ...
type HealthcheckEntity struct {
	sync.Mutex
	runningHealthchecks domain.ServiceInfoConf // TODO: map much better
	// memoryWorker       domain.MemoryWorker
	healthcheckChecker domain.HealthcheckChecker
	ipvsadm            domain.IPVSWorker
	dw                 domain.DummyWorker
	idGenerator        domain.IDgenerator
	logging            *logrus.Logger
	dummyEntities      map[string]*dummyEntity
}

type dummyEntity struct {
	totalForDummy      int
	announcedForDummy  int
	isAnnouncedAtDummy bool
}

// NewHealthcheckEntity ...
func NewHealthcheckEntity( // memoryWorker domain.MemoryWorker,
	healthcheckChecker domain.HealthcheckChecker,
	ipvsadm domain.IPVSWorker,
	dw domain.DummyWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HealthcheckEntity {

	return &HealthcheckEntity{
		runningHealthchecks: domain.ServiceInfoConf{}, // need for append
		// memoryWorker:       memoryWorker,
		healthcheckChecker: healthcheckChecker,
		ipvsadm:            ipvsadm,
		dw:                 dw,
		idGenerator:        idGenerator,
		logging:            logging,
		dummyEntities:      make(map[string]*dummyEntity),
	}
}

// NewServiceToHealthchecks - add service for healthchecks
func (hc *HealthcheckEntity) NewServiceToHealthchecks(newHCService *domain.ServiceInfo, id string) error {
	hc.Lock()
	defer hc.Unlock()

	if _, inMap := hc.runningHealthchecks[newHCService.Address]; inMap {
		return fmt.Errorf("new service to healthchecks error: service %v already exist, can't add it. Need to use update", newHCService.Address)
	}
	hc.runningHealthchecks[newHCService.Address] = newHCService

	// check we have same service ip in dummy map
	if _, inMap := hc.dummyEntities[newHCService.IP]; !inMap {
		// that also set announcedForDummy to 0 and isAnnouncedAtDummy to false
		hc.dummyEntities[newHCService.IP] = &dummyEntity{totalForDummy: 1}
	} else {
		hc.dummyEntities[newHCService.IP].totalForDummy++
	}

	if err := hc.addServiceToIPVS(newHCService, id); err != nil {
		return fmt.Errorf("can't add srvice to IPVS: %v", err)
	}
	go hc.startFirstChecksForService(newHCService, id)
	return nil
}

// RemoveServiceFromHealthchecks will work until removed
func (hc *HealthcheckEntity) RemoveServiceFromHealthchecks(removeHCService *domain.ServiceInfo, id string) error {
	hc.Lock()
	if _, inMap := hc.runningHealthchecks[removeHCService.Address]; !inMap {
		hc.Unlock()
		return fmt.Errorf("critical error: somehow key %v not in map %v\n", removeHCService.Address, hc.runningHealthchecks)
	}
	hc.Unlock()
	hc.runningHealthchecks[removeHCService.Address].HCStop <- struct{}{}
	hc.logging.Tracef("send stop checks for service %v", removeHCService.Address)
	<-hc.runningHealthchecks[removeHCService.Address].HCStopped
	hc.Lock()
	defer hc.Unlock()
	// paranoid checks key start
	if _, inMap := hc.dummyEntities[removeHCService.IP]; !inMap {
		return fmt.Errorf("critical error: somehow key %v not in map %v", removeHCService.IP, hc.dummyEntities)
	}
	if _, inMap := hc.runningHealthchecks[removeHCService.Address]; !inMap {
		return fmt.Errorf("critical error: somehow key %v not in map %v\n", removeHCService.Address, hc.runningHealthchecks)
	}
	// paranoid checks key end

	// if last service in map
	if hc.dummyEntities[removeHCService.IP].totalForDummy <= 1 {
		if hc.dummyEntities[removeHCService.IP].isAnnouncedAtDummy {
			if err := hc.dw.RemoveFromDummy(removeHCService.IP, id); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":   healthcheckName,
					"event id": id,
				}).Errorf("remove from dummy fail: %v", err)
			} else {
				hc.logging.WithFields(logrus.Fields{
					"entity":   healthcheckName,
					"event id": id,
				}).Infof("removed announce %v from dummy beacose service removed", removeHCService.IP)
			}
		}
		delete(hc.dummyEntities, removeHCService.IP)
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Infof("removed service %v from dummy entities beacose service removed", removeHCService.IP)
	} else {
		// if service up they in map ip announce to dummy
		if hc.runningHealthchecks[removeHCService.Address].IsUp {
			// don't remove announce here, or do some checks, because we do that only in announce logic func
			hc.dummyEntities[removeHCService.IP].announcedForDummy--
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": id,
			}).Infof("decrease announced to dummy for %v\n", removeHCService.IP)
		}
		// service not last, so we decrease total values
		hc.dummyEntities[removeHCService.IP].totalForDummy--
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Infof("decrease total for dummy for %v\n", removeHCService.IP)
	}
	delete(hc.runningHealthchecks, removeHCService.Address)
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": id,
	}).Infof("remove %v from running healthchecks", removeHCService.Address)

	if err := hc.removeServiceFromIPVS(removeHCService, id); err != nil {
		return fmt.Errorf("can't remove service from IPVS: %v", err)
	} else {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Infof("remove %v from ipvs", removeHCService.Address)
	}
	hc.logging.Infof("get checks stopped from service %v", removeHCService.Address)
	return nil
}

// UpdateServiceAtHealthchecks ...
func (hc *HealthcheckEntity) UpdateServiceAtHealthchecks(updateHCService *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	hc.Lock()
	if _, inMap := hc.runningHealthchecks[updateHCService.Address]; !inMap {
		hc.Unlock()
		return nil, fmt.Errorf("critical error: somehow key %v not in map %v\n", updateHCService.Address, hc.runningHealthchecks)
	}

	currentApplicationServers := getCopyOfApplicationServersFromService(hc.runningHealthchecks[updateHCService.Address]) // copy of current services
	hc.Unlock()

	hc.logging.Tracef("get update service job, sending stop checks for service %v", updateHCService.Address)
	hc.runningHealthchecks[updateHCService.Address].HCStop <- struct{}{}
	<-hc.runningHealthchecks[updateHCService.Address].HCStopped
	hc.logging.Tracef("healthchecks stopped for update service job %v", updateHCService.Address)
	hc.Lock()
	hc.enrichApplicationServersHealthchecks(updateHCService, currentApplicationServers, hc.runningHealthchecks[updateHCService.Address].IsUp) // lock updateHCService
	hc.Unlock()
	_, _, applicationServersForRemove := formDiffApplicationServers(updateHCService.ApplicationServers, currentApplicationServers)
	if len(applicationServersForRemove) != 0 {
		for _, k := range applicationServersForRemove {
			if err := hc.excludeApplicationServerFromIPVS(updateHCService, currentApplicationServers[k], id); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":   healthcheckName,
					"event id": id,
				}).Errorf("Healthcheck error: exclude application server from IPVS: %v", err)
			}
		}
	}
	hc.Lock()
	defer hc.Unlock()
	hc.runningHealthchecks[updateHCService.Address] = updateHCService

	go hc.startFirstChecksForService(updateHCService, id) // TODO can write it better. that will not remove announce at time
	return updateHCService, nil
}

func (hc *HealthcheckEntity) enrichApplicationServersHealthchecks(newServiceHealthcheck *domain.ServiceInfo, oldApplicationServers domain.ApplicationServers, oldIsUpState bool) {
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

		if _, isFunded := oldApplicationServers[k]; isFunded {
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

func (hc *HealthcheckEntity) GetServiceState(hcService *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	hc.Lock()
	defer hc.Unlock()
	for _, runHC := range hc.runningHealthchecks {
		if runHC.Address == hcService.Address {
			return copyServiceInfo(runHC), nil
		}
	}
	return nil, fmt.Errorf("get service state fail: service %v not found in healthchecks", hcService.Address)
}

func (hc *HealthcheckEntity) GetServicesState(id string) (domain.ServiceInfoConf, error) {
	hc.Lock()
	defer hc.Unlock()
	if len(hc.runningHealthchecks) == 0 {
		return nil, fmt.Errorf("active hc services: %v", len(hc.runningHealthchecks))
	}
	copyOfServiceInfos := make(domain.ServiceInfoConf, len(hc.runningHealthchecks))
	for i, runHC := range hc.runningHealthchecks {
		copyOfServiceInfos[i] = copyServiceInfo(runHC)
	}
	return copyOfServiceInfos, nil
}
