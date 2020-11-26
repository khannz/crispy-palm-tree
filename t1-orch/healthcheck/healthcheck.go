package healthcheck

// TODO: hc to domain module
import (
	"fmt"
	"sync"
	"time"

	domain "github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"
const protocolICMP = 1

type failedApplicationServers struct { // TODO: remove that struct
	sync.Mutex
	wg    *sync.WaitGroup
	count int
}

// HeathcheckEntity ...
type HeathcheckEntity struct {
	sync.Mutex
	runningHeathchecks []*domain.ServiceInfo // TODO: map much better
	memoryWorker       domain.MemoryWorker
	ipvsadm            domain.IPVSWorker
	dw                 domain.DummyWorker
	idGenerator        domain.IDgenerator
	logging            *logrus.Logger
	announcedServices  map[string]int
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(memoryWorker domain.MemoryWorker,
	ipvsadm domain.IPVSWorker,
	dw domain.DummyWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HeathcheckEntity {

	return &HeathcheckEntity{
		runningHeathchecks: []*domain.ServiceInfo{}, // need for append
		memoryWorker:       memoryWorker,
		ipvsadm:            ipvsadm,
		dw:                 dw,
		idGenerator:        idGenerator,
		logging:            logging,
		announcedServices:  make(map[string]int),
	}
}

func (hc *HeathcheckEntity) addNewServiceToMayAnnouncedServices(serviceIP string) {
	if _, inMap := hc.announcedServices[serviceIP]; inMap {
		hc.announcedServices[serviceIP]++
		return
	}
	hc.announcedServices[serviceIP] = 0 // add new service to annonced pool
}

func (hc *HeathcheckEntity) addServiceToIPVS(hcService *domain.ServiceInfo, id string) error {
	vip, port, routingType, balanceType, protocol, err := PrepareServiceForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol)
	if err != nil {
		return fmt.Errorf("error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.NewIPVSService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		id); err != nil {
		return fmt.Errorf("error when ipvsadm create service: %v", err)
	}

	return nil
}

// NewServiceToHealtchecks - add service for healthchecks
func (hc *HeathcheckEntity) NewServiceToHealtchecks(hcService *domain.ServiceInfo, id string) error {
	hc.Lock()
	defer hc.Unlock()

	hc.enrichApplicationServersHealthchecks(hcService, nil, false) // lock hcService
	hc.runningHeathchecks = append(hc.runningHeathchecks, hcService)
	hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
	if err := hc.addServiceToIPVS(hcService, id); err != nil {
		return fmt.Errorf("can't add srvice to IPVS: %v", err)
	}
	go hc.startHealthchecksForCurrentService(hcService)
	return nil
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

func (hc *HeathcheckEntity) removeServiceFromIPVS(hcService *domain.ServiceInfo, id string) error {
	vip, port, _, _, protocol, err := PrepareServiceForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol)
	if err != nil {
		return fmt.Errorf("error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.RemoveIPVSService(vip,
		port,
		protocol,
		id); err != nil {
		return fmt.Errorf("error when ipvsadm remove service: %v", err)
	}
	return nil
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
		// do not include new app servers here!
		_, _, applicationServersForRemove := formDiffApplicationServers(hcService.ApplicationServers, currentApplicationServers)
		if len(applicationServersForRemove) != 0 {
			for _, k := range applicationServersForRemove {
				// if hc.isApplicationServerInIPSVService(hcService.IP,
				// 	hcService.Port,
				// 	currentApplicationServers[k].IP,
				// 	currentApplicationServers[k].Port,
				// 	id) {
				if err := hc.excludeApplicationServerFromIPVS(hcService, currentApplicationServers[k], id); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": id,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
					// }
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
			HCType:           applicationServer.InternalHC.HCType,
			HCAddress:        applicationServer.InternalHC.HCAddress,
			HCTimeout:        applicationServer.InternalHC.HCTimeout,
			RetriesForUP:     applicationServer.InternalHC.RetriesForUP,
			RetriesForDown:   applicationServer.InternalHC.RetriesForDown,
			LastIndexForUp:   applicationServer.InternalHC.LastIndexForUp,
			LastIndexForDown: applicationServer.InternalHC.LastIndexForDown,
			NearFieldsMode:   applicationServer.InternalHC.NearFieldsMode,
			UserDefinedData:  applicationServer.InternalHC.UserDefinedData,
		}

		applicationServers[k] = &domain.ApplicationServer{
			Address:    applicationServer.Address,
			IP:         applicationServer.IP,
			Port:       applicationServer.Port,
			IsUp:       applicationServer.IsUp,
			HCAddress:  applicationServer.HCAddress,
			InternalHC: tmpApplicationServerInternal,
		}
	}
	return applicationServers
}

func (hc *HeathcheckEntity) enrichApplicationServersHealthchecks(newServiceHealthcheck *domain.ServiceInfo, oldApplicationServers map[string]*domain.ApplicationServer, oldIsUpState bool) {
	newServiceHealthcheck.Lock()
	defer newServiceHealthcheck.Unlock()
	newServiceHealthcheck.IsUp = oldIsUpState
	newServiceHealthcheck.HCStop = make(chan struct{}, 1)
	newServiceHealthcheck.HCStopped = make(chan struct{}, 1)

	for k := range newServiceHealthcheck.ApplicationServers {
		internalHC := domain.InternalHC{}
		internalHC.HCType = newServiceHealthcheck.HCType
		internalHC.HCAddress = newServiceHealthcheck.ApplicationServers[k].HCAddress
		internalHC.HCTimeout = newServiceHealthcheck.HCTimeout
		internalHC.LastIndexForUp = 0
		internalHC.LastIndexForDown = 0
		internalHC.NearFieldsMode = newServiceHealthcheck.HCNearFieldsMode
		internalHC.UserDefinedData = newServiceHealthcheck.HCUserDefinedData

		retriesCounterForUp := make([]bool, newServiceHealthcheck.HCRetriesForUP)
		retriesCounterForDown := make([]bool, newServiceHealthcheck.HCRetriesForDown)

		if _, isFinded := oldApplicationServers[k]; isFinded {
			fillNewBooleanArray(retriesCounterForUp, oldApplicationServers[k].InternalHC.RetriesForUP)
			fillNewBooleanArray(retriesCounterForDown, oldApplicationServers[k].InternalHC.RetriesForDown)
			internalHC.RetriesForUP = retriesCounterForUp
			internalHC.RetriesForDown = retriesCounterForDown
			newServiceHealthcheck.ApplicationServers[k].IsUp = oldApplicationServers[k].IsUp
			newServiceHealthcheck.ApplicationServers[k].InternalHC = internalHC
			hc.logging.Debugf("application server %v was found, is up state was moved", newServiceHealthcheck.ApplicationServers[k].Address)
			continue
		}
		internalHC.RetriesForUP = retriesCounterForUp
		internalHC.RetriesForDown = retriesCounterForDown
		newServiceHealthcheck.ApplicationServers[k].IsUp = false
		newServiceHealthcheck.ApplicationServers[k].InternalHC = internalHC
		hc.logging.Debugf("application server %v NOT found, is up state set false", newServiceHealthcheck.ApplicationServers[k].Address)
	}
}

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

func (hc *HeathcheckEntity) startHealthchecksForCurrentService(hcService *domain.ServiceInfo) {
	// first run hc at create entity
	idForCheckService := hc.idGenerator.NewID()
	hc.CheckApplicationServersInService(hcService, idForCheckService) // lock hc, hcService, dummy
	hc.logging.Infof("hc service: %v", hcService)
	ticker := time.NewTicker(hcService.HCRepeat)
	for {
		select {
		case <-hcService.HCStop:
			hc.logging.Tracef("get stop checks command for service %v; send checks stoped and return", hcService.Address)
			hcService.HCStopped <- struct{}{}
			return
		case <-ticker.C:
			hc.CheckApplicationServersInService(hcService, idForCheckService) // lock hc, hcService, dummy
		}
	}
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(hcService *domain.ServiceInfo, id string) {
	fs := &failedApplicationServers{wg: new(sync.WaitGroup)} // TODO: move wg to ServiceInfo

	for k := range hcService.ApplicationServers {
		fs.wg.Add(1)
		go hc.checkApplicationServerInService(hcService,
			fs,
			k,
			id) // lock hcService
	}
	fs.wg.Wait()
	percentageUp := percentageOfUp(len(hcService.ApplicationServers), fs.count)
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": id,
	}).Debugf("Heathcheck: in service %v failed services is %v of %v; %v up percent of %v max for this service",
		hcService.Address,
		fs.count,
		len(hcService.ApplicationServers),
		percentageUp,
		hcService.AlivedAppServersForUp)
	isServiceUp := percentageOfDownBelowMPercentOfAlivedForUp(percentageUp, hcService.AlivedAppServersForUp)
	hc.logging.Tracef("Old service state %v. New service state %v", hcService.IsUp, isServiceUp)

	if !hcService.IsUp && isServiceUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Warnf("service %v is up now", hcService.Address)
		hcService.IsUp = true
		hc.annonceLogic(hcService.IP, hcService.IsUp, id) // lock hc and dummy
	} else if hcService.IsUp && !isServiceUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Warnf("service %v is down now", hcService.Address)
		hcService.IsUp = false
		hc.annonceLogic(hcService.IP, hcService.IsUp, id) // lock hc and dummy
	} else {
		hc.logging.Debugf("service state not changed: is up: %v", hcService.IsUp)
	}
	hc.updateInStorage(hcService, id)
}

// annonceLogic - used when service change state
func (hc *HeathcheckEntity) annonceLogic(serviceIP string, newIsUpServiceState bool, id string) {
	hc.Lock()
	defer hc.Unlock()
	isServiceAnoncedNow := hc.announcedServices[serviceIP] > 0
	//
	if isServiceAnoncedNow {
		if newIsUpServiceState {
			hc.addNewServiceToMayAnnouncedServices(serviceIP)
			// log error
			return
		}
		// isServiceAnoncedNow && !newIsUpServiceState
		if i, inMap := hc.announcedServices[serviceIP]; inMap {
			hc.announcedServices[serviceIP] = i - 1
			if hc.announcedServices[serviceIP] == 0 {
				if err := hc.dw.RemoveFromDummy(serviceIP, id); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": id,
					}).Errorf("remove from dummy fail: %v", err)
				}
				return
			}
		}
		// log error
		return
	}
	// !isServiceAnoncedNow
	if newIsUpServiceState {
		if i, inMap := hc.announcedServices[serviceIP]; inMap {
			hc.announcedServices[serviceIP] = i + 1 // set 1, i=0 here
			if err := hc.dw.AddToDummy(serviceIP, id); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":   healthcheckName,
					"event id": id,
				}).Errorf("add to dummy fail: %v", err)
			}
			return
		}
		// log error
		return
	}
	// !isServiceAnoncedNow && !newIsUpServiceState return
}

func (hc *HeathcheckEntity) updateInStorage(hcService *domain.ServiceInfo, id string) {
	if err := hc.memoryWorker.UpdateService(hcService); err != nil {
		hc.logging.Warnf("can't update service %v from healtchecks: %v",
			hcService.Address,
			err)
	}

}

func (hc *HeathcheckEntity) isApplicationServerOkNow(hcService *domain.ServiceInfo,
	fs *failedApplicationServers,
	applicationServerInfoKey string,
	id string) bool {
	// FIXME: set fwmark
	switch hcService.HCType {
	case "tcp":
		return IsTcpCheckOk(hcService.ApplicationServers[applicationServerInfoKey].HCAddress,
			hcService.HCTimeout,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)

	case "http": // FIXME: https checks here, no support for http
		return IsHttpsCheckOk(hcService.ApplicationServers[applicationServerInfoKey].HCAddress,
			hcService.HCTimeout,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	case "http-advanced":
		return IsHttpAdvancedCheckOk(hcService.HCType,
			hcService.ApplicationServers[applicationServerInfoKey].HCAddress,
			hcService.HCNearFieldsMode,
			hcService.HCUserDefinedData,
			hcService.HCTimeout,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	case "icmp":
		seq := 1
		// 	echoRequestSeqId.Lock() // FIXME: select sequence
		// if echoRequestSeqId.seq == 65535 {
		// 	echoRequestSeqId.seq = 1
		// } else {
		// 	echoRequestSeqId.seq++
		// }
		return IsIcmpCheckOk(hcService.ApplicationServers[applicationServerInfoKey].HCAddress,
			seq,
			hcService.HCTimeout,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Errorf("Heathcheck error: unknown healtcheck type: %v", hcService.HCType)
		return false // must never will bfe. all data already validated
	}
}

func (hc *HeathcheckEntity) checkApplicationServerInService(hcService *domain.ServiceInfo,
	fs *failedApplicationServers,
	applicationServerInfoKey string,
	id string) {
	// TODO: still can be refactored
	defer fs.wg.Done()
	isCheckOk := hc.isApplicationServerOkNow(hcService, fs, applicationServerInfoKey, id)
	hc.logging.Tracef("is check ok for server %v: %v",
		hcService.ApplicationServers[applicationServerInfoKey].Address,
		isCheckOk)

	hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoKey, isCheckOk)                                                     // lock hcService
	isApplicationServerUp, isApplicationServerChangeState := hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoKey, id) // lock hcService
	hc.logging.Tracef("for server %v:isApplicationServerUp: %v,isApplicationServerChangeState: %v ",
		hcService.ApplicationServers[applicationServerInfoKey].Address,
		isApplicationServerUp,
		isApplicationServerChangeState)

	if !isCheckOk {
		hc.logging.Debugf("one hc for application server %v is up: %v; is server change state: %v",
			hcService.ApplicationServers[applicationServerInfoKey].Address,
			isApplicationServerUp,
			isApplicationServerChangeState)
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if isApplicationServerChangeState {
				if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.ApplicationServers[applicationServerInfoKey], id); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": id,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
		}
		return
	}

	hc.logging.Debugf("one hc for application server %v is up: %v; is server change state: %v",
		hcService.ApplicationServers[applicationServerInfoKey].Address,
		isApplicationServerUp,
		isApplicationServerChangeState)
	if !isApplicationServerUp {
		fs.Lock()
		fs.count++
		fs.Unlock()
	}

	if isApplicationServerUp && isApplicationServerChangeState {
		if err := hc.inclideApplicationServerInIPVS(hcService, hcService.ApplicationServers[applicationServerInfoKey], id); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": id,
			}).Errorf("Heathcheck error: inclide application server in IPVS error: %v", err)
		}
		return
	}
}

func (hc *HeathcheckEntity) isApplicationServerUpAndStateChange(hcService *domain.ServiceInfo,
	applicationServerInfoKey string,
	id string) (bool, bool) {
	//return isUp and isChagedState booleans
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("real: %v, RetriesCounterForDown: %v", hcService.ApplicationServers[applicationServerInfoKey].Address, hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForUP)
	hc.logging.Tracef("real: %v, RetriesCounterForUp: %v", hcService.ApplicationServers[applicationServerInfoKey].Address, hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForDown)

	if hcService.ApplicationServers[applicationServerInfoKey].IsUp {
		// check it not down
		for _, isUp := range hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForDown {
			if isUp {
				return true, false // do not change up state
			}
		}
		hcService.ApplicationServers[applicationServerInfoKey].IsUp = false // if all hc fail at RetriesCounterForDown - change state
		hc.logging.WithFields(logrus.Fields{
			"event id": id,
		}).Warnf("at service %v real server %v DOWN", hcService.Address,
			hcService.ApplicationServers[applicationServerInfoKey].Address)
		return false, true
	}

	for _, isUp := range hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForUP {
		if !isUp {
			// do not change down state
			return false, false
		}
	}

	// all RetriesCounterForUp true
	hcService.ApplicationServers[applicationServerInfoKey].IsUp = true // if all hc fail at RetriesCounterForDown - change state
	hc.logging.WithFields(logrus.Fields{
		"event id": id,
	}).Warnf("at service %v real server %v UP", hcService.Address,
		hcService.ApplicationServers[applicationServerInfoKey].Address)
	return true, true

}

func (hc *HeathcheckEntity) moveApplicationServerStateIndexes(hcService *domain.ServiceInfo, applicationServerInfoKey string, isUpNow bool) {
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("moveApplicationServerStateIndexes: app srv index: %v,isUpNow: %v, total app srv at service: %v, RetriesForUP array len: %v, RetriesForDown array len: %v",
		applicationServerInfoKey,
		isUpNow,
		len(hcService.ApplicationServers),
		len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForUP),
		len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForDown))

	if len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForUP) < hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForUp+1 {
		hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForUp = 0
	}
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForUP[hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForUp] = isUpNow
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForUp++

	if len(hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForDown) < hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDown+1 {
		hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDown = 0
	}
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.RetriesForDown[hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDown] = isUpNow
	hcService.ApplicationServers[applicationServerInfoKey].InternalHC.LastIndexForDown++
}

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

func (hc *HeathcheckEntity) excludeApplicationServerFromIPVS(hcService *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer,
	id string) error {
	aS := map[string]*domain.ApplicationServer{applicationServer.IP + ":" + applicationServer.Port: applicationServer}
	vip, port, routingType, balanceType, protocol, applicationServers, err := PrepareDataForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol,
		aS)
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.RemoveIPVSApplicationServersFromService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id); err != nil {
		return fmt.Errorf("Error when ipvsadm remove application servers from service: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(hcService *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer,
	id string) error {
	aS := map[string]*domain.ApplicationServer{applicationServer.IP + ":" + applicationServer.Port: applicationServer}
	vip, port, routingType, balanceType, protocol, applicationServers, err := PrepareDataForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol,
		aS)
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}

	applicationServerPort, err := stringToUINT16(applicationServer.Port)
	if err != nil {
		return fmt.Errorf("can't convert port to uint16: %v", err)
	}

	oneAppSrvMap := make(map[string]uint16, 1)
	oneAppSrvMap[applicationServer.IP] = applicationServerPort
	// isApplicationServerInService, err := hc.ipvsadm.IsIPVSApplicationServerInService(vip,
	// 	port,
	// 	oneAppSrvMap,
	// 	"tmp fale id")
	// if err != nil {
	// 	return fmt.Errorf("can't check is application server in service: %v", err)
	// }
	// if !isApplicationServerInService {
	if err = hc.ipvsadm.AddIPVSApplicationServersForService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id); err != nil {
		return fmt.Errorf("Error when ipvsadm add application servers for service: %v", err)
	}
	// }
	return nil
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

func copyServiceInfo(hcService *domain.ServiceInfo) *domain.ServiceInfo {
	copyOfApplicationServersFromService := getCopyOfApplicationServersFromService(hcService)
	return &domain.ServiceInfo{
		Address:               hcService.Address,
		IP:                    hcService.IP,
		Port:                  hcService.Port,
		IsUp:                  hcService.IsUp,
		BalanceType:           hcService.BalanceType,
		RoutingType:           hcService.RoutingType,
		Protocol:              hcService.Protocol,
		AlivedAppServersForUp: hcService.AlivedAppServersForUp,
		HCType:                hcService.HCType,
		HCRepeat:              hcService.HCRepeat,
		HCTimeout:             hcService.HCTimeout,
		HCNearFieldsMode:      hcService.HCNearFieldsMode,
		HCUserDefinedData:     hcService.HCUserDefinedData,
		HCRetriesForUP:        hcService.HCRetriesForUP,
		HCRetriesForDown:      hcService.HCRetriesForDown,
		ApplicationServers:    copyOfApplicationServersFromService,
		HCStop:                make(chan struct{}, 1),
		HCStopped:             make(chan struct{}, 1),
	}
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
