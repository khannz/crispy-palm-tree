package portadapter

import (
	"fmt"
	"net"
	"sync"
	"time"

	domain "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"
const healthcheckID = "00000000-0000-0000-0000-000000000004"
const protocolICMP = 1

type failedApplicationServers struct { // TODO: remove that struct
	sync.Mutex
	wg    *sync.WaitGroup
	count int
}

type dummyWorker struct {
	sync.Mutex
}

// HeathcheckEntity ...
type HeathcheckEntity struct {
	sync.Mutex
	runningHeathchecks []*domain.HCService
	cacheStorage       domain.StorageActions
	ipvsadm            domain.IPVSWorker
	techInterface      *net.TCPAddr
	locker             *domain.Locker
	dw                 *dummyWorker
	isMockMode         bool
	logging            *logrus.Logger
	announcedServices  map[string]int
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(cacheStorage *StorageEntity,
	ipvsadm domain.IPVSWorker,
	rawTechInterface string,
	locker *domain.Locker,
	isMockMode bool,
	logging *logrus.Logger) *HeathcheckEntity {
	ti, _, _ := net.ParseCIDR(rawTechInterface + "/32")

	return &HeathcheckEntity{
		runningHeathchecks: []*domain.HCService{}, // need for append
		cacheStorage:       cacheStorage,
		ipvsadm:            ipvsadm,
		techInterface:      &net.TCPAddr{IP: ti},
		locker:             locker,
		dw:                 new(dummyWorker),
		isMockMode:         isMockMode,
		logging:            logging,
		announcedServices:  make(map[string]int),
	}
}

// UnknownDataStruct used for trying to get an unknown json or array of json's
type UnknownDataStruct struct {
	UnknowMap         map[string]string
	UnknowArrayOfMaps []map[string]string
}

func (hc *HeathcheckEntity) addNewServiceToMayAnnouncedServices(serviceIP string) bool {
	if _, inMap := hc.announcedServices[serviceIP]; inMap {
		return true // already in map
	}
	hc.announcedServices[serviceIP] = 0 // add new service to annonced pool
	return false
}

// StartHealthchecksForServices ...
func (hc *HeathcheckEntity) StartHealthchecksForServices(hcServices []*domain.HCService) error {
	hc.Lock()
	defer hc.Unlock()

	for _, hcService := range hcServices {
		hc.enrichApplicationServersHealthchecks(hcService, nil, false) // lock hcService
		hc.runningHeathchecks = append(hc.runningHeathchecks, hcService)
		alreadyAnnounced := hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
		if !alreadyAnnounced {
			if err := hc.addServiceToIPVS(hcService); err != nil {
				return fmt.Errorf("can't add srvice to IPVS: %v", err)
			}
		}
		go hc.startHealthchecksForCurrentService(hcService)
	}
	return nil
}

func (hc *HeathcheckEntity) addServiceToIPVS(hcService *domain.HCService) error {
	vip, port, routingType, balanceType, protocol, err := PrepareServiceForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol)
	if err != nil {
		return fmt.Errorf("error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.NewService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		nil,
		healthcheckID); err != nil {
		return fmt.Errorf("error when ipvsadm create service: %v", err)
	}
	return nil
}

// NewServiceToHealtchecks - add service for healthchecks
func (hc *HeathcheckEntity) NewServiceToHealtchecks(hcService *domain.HCService) error {
	hc.Lock()
	defer hc.Unlock()

	hc.enrichApplicationServersHealthchecks(hcService, nil, false) // lock hcService
	hc.runningHeathchecks = append(hc.runningHeathchecks, hcService)
	alreadyAnnounced := hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
	if !alreadyAnnounced {
		if err := hc.addServiceToIPVS(hcService); err != nil {
			return fmt.Errorf("can't add srvice to IPVS: %v", err)
		}
	}
	go hc.startHealthchecksForCurrentService(hcService)
	return nil
}

func (hc *HeathcheckEntity) removeServiceFromMayAnnouncedServices(serviceIP string) {
	_, isFinded := hc.announcedServices[serviceIP]
	if isFinded {
		if hc.announcedServices[serviceIP] == 0 {
			delete(hc.announcedServices, serviceIP)
		}
	}
}

// RemoveServiceFromHealtchecks will work until removed
func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(hcService *domain.HCService) error {
	hc.Lock()
	indexForRemove, isFinded := hc.findServiceInHealtcheckSlice(hcService.Address)
	hc.Unlock()
	if isFinded {
		hc.logging.Tracef("send stop checks for service %v", hcService.Address)
		hc.runningHeathchecks[indexForRemove].HCStop <- struct{}{}
		<-hc.runningHeathchecks[indexForRemove].HCStopped
		hc.annonceLogic(hcService.IP, false) // lock hc and dummy; may remove annonce
		hc.Lock()
		hc.removeServiceFromMayAnnouncedServices(hcService.IP) // hc must be locked
		hc.runningHeathchecks = append(hc.runningHeathchecks[:indexForRemove], hc.runningHeathchecks[indexForRemove+1:]...)
		hc.Unlock()
		if err := hc.removeServiceFromIPVS(hcService); err != nil {
			return fmt.Errorf("can't add srvice to IPVS: %v", err)
		}
		hc.logging.Tracef("get checks stopped from service %v", hcService.Address)
		return nil
	}

	return fmt.Errorf("Heathcheck error: RemoveServiceFromHealtchecks error: service %v not found",
		hcService.Address)
}

func (hc *HeathcheckEntity) removeServiceFromIPVS(hcService *domain.HCService) error {
	vip, port, _, _, protocol, err := PrepareServiceForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol)
	if err != nil {
		return fmt.Errorf("error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.RemoveService(vip,
		port,
		protocol,
		healthcheckID); err != nil {
		return fmt.Errorf("error when ipvsadm remove service: %v", err)
	}
	return nil
}

// UpdateServiceAtHealtchecks ...
func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(hcService *domain.HCService) (*domain.HCService, error) {
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
		applicationServersForRemove := formApplicationServersForRemove(hcService.HCApplicationServers, currentApplicationServers)
		if len(applicationServersForRemove) != 0 {
			for i := range applicationServersForRemove {
				if hc.isApplicationServerInIPSVService(hcService.IP,
					hcService.Port,
					applicationServersForRemove[i].IP,
					applicationServersForRemove[i].Port) {
					if err := hc.excludeApplicationServerFromIPVS(hcService, applicationServersForRemove[i]); err != nil {
						hc.logging.WithFields(logrus.Fields{
							"entity":   healthcheckName,
							"event id": healthcheckID,
						}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
					}
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

func formApplicationServersForRemove(newApplicationServers []*domain.HCApplicationServer, oldApplicationServers []*domain.HCApplicationServer) []*domain.HCApplicationServer {
	applicationServersForRemove := []*domain.HCApplicationServer{}
	for i := range oldApplicationServers {
		_, isFinded := findApplicationServer(oldApplicationServers[i].Address, newApplicationServers)
		if !isFinded {
			applicationServersForRemove = append(applicationServersForRemove, oldApplicationServers[i])
			continue
		}
	}
	return applicationServersForRemove
}

func getCopyOfApplicationServersFromService(serviceInfo *domain.HCService) []*domain.HCApplicationServer {
	applicationServers := make([]*domain.HCApplicationServer, len(serviceInfo.HCApplicationServers))
	for i, applicationServer := range serviceInfo.HCApplicationServers {
		tmpHCApplicationServer := domain.InternalHC{
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

		applicationServers[i] = &domain.HCApplicationServer{
			Address:             applicationServer.Address,
			IP:                  applicationServer.IP,
			Port:                applicationServer.Port,
			IsUp:                applicationServer.IsUp,
			HCAddress:           applicationServer.HCAddress,
			InternalHC:          tmpHCApplicationServer,
			ExampleBashCommands: applicationServer.ExampleBashCommands,
		}
	}
	return applicationServers
}

func (hc *HeathcheckEntity) enrichApplicationServersHealthchecks(newServiceHealthcheck *domain.HCService, oldApplicationServers []*domain.HCApplicationServer, oldIsUpState bool) {
	newServiceHealthcheck.Lock()
	defer newServiceHealthcheck.Unlock()
	newServiceHealthcheck.IsUp = oldIsUpState
	newServiceHealthcheck.HCStop = make(chan struct{}, 1)
	newServiceHealthcheck.HCStopped = make(chan struct{}, 1)

	for i := range newServiceHealthcheck.HCApplicationServers {
		internalHC := domain.InternalHC{}
		internalHC.HCType = newServiceHealthcheck.HCType
		internalHC.HCAddress = newServiceHealthcheck.HCApplicationServers[i].HCAddress
		internalHC.HCTimeout = newServiceHealthcheck.HCTimeout
		internalHC.LastIndexForUp = 0
		internalHC.LastIndexForDown = 0
		internalHC.NearFieldsMode = newServiceHealthcheck.HCNearFieldsMode
		internalHC.UserDefinedData = newServiceHealthcheck.HCUserDefinedData

		retriesCounterForUp := make([]bool, newServiceHealthcheck.HCRetriesForUP)
		retriesCounterForDown := make([]bool, newServiceHealthcheck.HCRetriesForDown)

		j, isFinded := findApplicationServer(newServiceHealthcheck.HCApplicationServers[i].Address, oldApplicationServers)
		if isFinded {
			fillNewBooleanArray(retriesCounterForUp, oldApplicationServers[j].InternalHC.RetriesForUP)
			fillNewBooleanArray(retriesCounterForDown, oldApplicationServers[j].InternalHC.RetriesForDown)
			internalHC.RetriesForUP = retriesCounterForUp
			internalHC.RetriesForDown = retriesCounterForDown
			newServiceHealthcheck.HCApplicationServers[i].IsUp = oldApplicationServers[j].IsUp
			newServiceHealthcheck.HCApplicationServers[i].InternalHC = internalHC
			hc.logging.Debugf("application server %v was found, is up state was moved", newServiceHealthcheck.HCApplicationServers[i].Address)
			continue
		}
		internalHC.RetriesForUP = retriesCounterForUp
		internalHC.RetriesForDown = retriesCounterForDown
		newServiceHealthcheck.HCApplicationServers[i].IsUp = false
		newServiceHealthcheck.HCApplicationServers[i].InternalHC = internalHC
		hc.logging.Debugf("application server %v NOT found, is up state set false", newServiceHealthcheck.HCApplicationServers[i].Address)
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

func (hc *HeathcheckEntity) startHealthchecksForCurrentService(hcService *domain.HCService) {
	// first run hc at create entity
	hc.CheckApplicationServersInService(hcService) // lock hc, hcService, dummy
	hc.logging.Infof("hc service: %v", hcService)
	ticker := time.NewTicker(hcService.HCRepeat)
	for {
		select {
		case <-hcService.HCStop:
			hc.logging.Tracef("get stop checks command for service %v; send checks stoped and return", hcService.Address)
			hcService.HCStopped <- struct{}{}
			return
		case <-ticker.C:
			hc.CheckApplicationServersInService(hcService) // lock hc, hcService, dummy
		}
	}
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(hcService *domain.HCService) {
	fs := &failedApplicationServers{wg: new(sync.WaitGroup)} // TODO: move wg to HCService

	for i := range hcService.HCApplicationServers {
		fs.wg.Add(1)
		go hc.checkApplicationServerInService(hcService,
			fs,
			i) // lock hcService
	}
	fs.wg.Wait()
	percentageUp := percentageOfUp(len(hcService.HCApplicationServers), fs.count)
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": healthcheckID,
	}).Debugf("Heathcheck: in service %v failed services is %v of %v; %v up percent of %v max for this service",
		hcService.Address,
		fs.count,
		len(hcService.HCApplicationServers),
		percentageUp,
		hcService.AlivedAppServersForUp)
	isServiceUp := percentageOfDownBelowMPercentOfAlivedForUp(percentageUp, hcService.AlivedAppServersForUp)
	hc.logging.Tracef("Old service state %v. New service state %v", hcService.IsUp, isServiceUp)

	if !hcService.IsUp && isServiceUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Warnf("service %v is up now", hcService.Address)
		hcService.IsUp = true
		hc.annonceLogic(hcService.IP, hcService.IsUp) // lock hc and dummy
	} else if hcService.IsUp && !isServiceUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Warnf("service %v is down now", hcService.Address)
		hcService.IsUp = false
		hc.annonceLogic(hcService.IP, hcService.IsUp) // lock hc and dummy
	} else {
		hc.logging.Debugf("service state not changed: is up: %v", hcService.IsUp)
	}
	hc.updateInStorage(hcService)
}

// annonceLogic - used when service change state
func (hc *HeathcheckEntity) annonceLogic(serviceIP string, newIsUpServiceState bool) {
	hc.Lock()
	defer hc.Unlock()
	isServiceAnoncedNow := hc.announcedServices[serviceIP] > 0
	//
	if isServiceAnoncedNow {
		if newIsUpServiceState {
			if i, inMap := hc.announcedServices[serviceIP]; inMap {
				hc.announcedServices[serviceIP] = i + 1
				return
			}
			// log error
			return
		}
		// isServiceAnoncedNow && !newIsUpServiceState
		if i, inMap := hc.announcedServices[serviceIP]; inMap {
			hc.announcedServices[serviceIP] = i - 1
			if hc.announcedServices[serviceIP] == 0 {
				hc.removeFromDummyWrapper(serviceIP)
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
			hc.addToDummyWrapper(serviceIP)
			return
		}
		// log error
		return
	}
	// !isServiceAnoncedNow && !newIsUpServiceState return
}

func (hc *HeathcheckEntity) updateInStorage(hcService *domain.HCService) {
	errUpdataCache := hc.cacheStorage.UpdateHCService(hcService, healthcheckID)
	if errUpdataCache != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck update info in cache fail: %v", errUpdataCache)
	}
}

func (hc *HeathcheckEntity) isApplicationServerOkNow(hcService *domain.HCService,
	fs *failedApplicationServers,
	applicationServerInfoIndex int) bool {
	switch hcService.HCType {
	case "tcp":
		return hc.tcpCheckOk(hcService.HCApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCTimeout)
	case "http":
		return hc.httpCheckOk(hcService.HCApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCTimeout)
	case "http-advanced":
		return hc.httpAdvancedCheckOk(hcService.HCType,
			hcService.HCApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCNearFieldsMode,
			hcService.HCUserDefinedData,
			hcService.HCTimeout)
	case "icmp":
		return hc.icmpCheckOk(hcService.HCApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCTimeout)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck error: unknown healtcheck type: %v", hcService.HCType)
		return false // must never will bfe. all data already validated
	}
}

func (hc *HeathcheckEntity) checkApplicationServerInService(hcService *domain.HCService,
	fs *failedApplicationServers,
	applicationServerInfoIndex int) {
	// TODO: still can be refactored
	defer fs.wg.Done()
	var isApplicationServerUp, isApplicationServerChangeState bool
	isCheckOk := hc.isApplicationServerOkNow(hcService, fs, applicationServerInfoIndex)

	hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
	isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
	if !isCheckOk {
		hc.logging.Debugf("one hc for application server %v fail: %v; is change state: %v",
			hcService.HCApplicationServers[applicationServerInfoIndex].Address,
			isApplicationServerUp,
			isApplicationServerChangeState)
		if !isApplicationServerUp {
			if isApplicationServerChangeState {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.HCApplicationServers[applicationServerInfoIndex]); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": healthcheckID,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
		}
		return
	}

	hc.logging.Debugf("one hc for application server %v ok: %v; is change state: %v",
		hcService.HCApplicationServers[applicationServerInfoIndex].Address,
		isApplicationServerUp,
		isApplicationServerChangeState)
	if !isApplicationServerUp {
		fs.Lock()
		fs.count++
		fs.Unlock()
	}

	if isApplicationServerUp && isApplicationServerChangeState {
		if err := hc.inclideApplicationServerInIPVS(hcService, hcService.HCApplicationServers[applicationServerInfoIndex]); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": healthcheckID,
			}).Errorf("Heathcheck error: inclide application server in IPVS error: %v", err)
		}
		return
	}
}

func (hc *HeathcheckEntity) isApplicationServerUpAndStateChange(hcService *domain.HCService,
	applicationServerInfoIndex int) (bool, bool) {
	//return isUp and isChagedState booleans
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("real: %v, RetriesCounterForDown: %v", hcService.HCApplicationServers[applicationServerInfoIndex].Address, hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForUP)
	hc.logging.Tracef("real: %v, RetriesCounterForUp: %v", hcService.HCApplicationServers[applicationServerInfoIndex].Address, hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForDown)

	if hcService.HCApplicationServers[applicationServerInfoIndex].IsUp {
		// check it not down
		for _, isUp := range hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForDown {
			if isUp {
				return true, false // do not change up state
			}
		}
		hcService.HCApplicationServers[applicationServerInfoIndex].IsUp = false // if all hc fail at RetriesCounterForDown - change state
		hc.logging.WithFields(logrus.Fields{
			"event id": healthcheckID,
		}).Warnf("at service %v real server %v DOWN", hcService.Address,
			hcService.HCApplicationServers[applicationServerInfoIndex].Address)
		return false, true
	}

	for _, isUp := range hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForUP {
		if !isUp {
			// do not change down state
			return false, false
		}
	}

	// all RetriesCounterForUp true
	hcService.HCApplicationServers[applicationServerInfoIndex].IsUp = true // if all hc fail at RetriesCounterForDown - change state
	hc.logging.WithFields(logrus.Fields{
		"event id": healthcheckID,
	}).Warnf("at service %v real server %v UP", hcService.Address,
		hcService.HCApplicationServers[applicationServerInfoIndex].Address)
	return true, true

}

func (hc *HeathcheckEntity) moveApplicationServerStateIndexes(hcService *domain.HCService, applicationServerInfoIndex int, isUpNow bool) {
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("moveApplicationServerStateIndexes: app srv index: %v,isUpNow: %v, total app srv at service: %v, RetriesForUP array len: %v, RetriesForDown array len: %v",
		applicationServerInfoIndex,
		isUpNow,
		len(hcService.HCApplicationServers),
		len(hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForUP),
		len(hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForDown))

	if len(hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForUP) < hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForUp+1 {
		hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForUp = 0
	}
	hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForUP[hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForUp] = isUpNow
	hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForUp++

	if len(hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForDown) < hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForDown+1 {
		hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForDown = 0
	}
	hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.RetriesForDown[hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForDown] = isUpNow
	hcService.HCApplicationServers[applicationServerInfoIndex].InternalHC.LastIndexForDown++
}

// TODO: much faster if we have func isApplicationServersInIPSVSerrvice, return []int (indexes for app srv not fount)
func (hc *HeathcheckEntity) isApplicationServerInIPSVService(hcServiceIP, rawHcServicePort, applicationServerIP, rawApplicationServerPort string) bool {
	hcServicePort, err := stringToUINT16(rawHcServicePort)
	if err != nil {
		hc.logging.Errorf("can't convert port to uint16: %v", err)
	}
	applicationServerPort, err := stringToUINT16(rawApplicationServerPort)
	if err != nil {
		hc.logging.Errorf("can't convert port to uint16: %v", err)
	}
	isApplicationServerInService, err := hc.ipvsadm.IsApplicationServerInService(hcServiceIP,
		hcServicePort,
		applicationServerIP,
		applicationServerPort)
	if err != nil {
		hc.logging.Errorf("can't check is application server in service: %v", err)
		return false
	}
	return isApplicationServerInService
}

func (hc *HeathcheckEntity) excludeApplicationServerFromIPVS(hcService *domain.HCService,
	applicationServer *domain.HCApplicationServer) error {
	vip, port, routingType, balanceType, protocol, applicationServers, err := PrepareDataForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol,
		[]*domain.HCApplicationServer{applicationServer})
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.RemoveApplicationServersFromService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		healthcheckID); err != nil {
		return fmt.Errorf("Error when ipvsadm remove application servers from service: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(hcService *domain.HCService,
	applicationServer *domain.HCApplicationServer) error {
	// !!!

	vip, port, routingType, balanceType, protocol, applicationServers, err := PrepareDataForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol,
		[]*domain.HCApplicationServer{applicationServer})
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}

	applicationServerPort, err := stringToUINT16(applicationServer.Port)
	if err != nil {
		return fmt.Errorf("can't convert port to uint16: %v", err)
	}
	isApplicationServerInService, err := hc.ipvsadm.IsApplicationServerInService(vip,
		port,
		applicationServer.IP,
		applicationServerPort)
	if err != nil {
		return fmt.Errorf("can't check is application server in service: %v", err)
	}
	if !isApplicationServerInService {
		if err = hc.ipvsadm.AddApplicationServersForService(vip,
			port,
			routingType,
			balanceType,
			protocol,
			applicationServers,
			healthcheckID); err != nil {
			return fmt.Errorf("Error when ipvsadm add application servers for service: %v", err)
		}
	}
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

func findApplicationServer(address string, oldApplicationServers []*domain.HCApplicationServer) (int, bool) {
	var findedIndex int
	var isFinded bool
	if oldApplicationServers == nil {
		return findedIndex, isFinded
	}
	for index, oldApplicationServer := range oldApplicationServers {
		if address == oldApplicationServer.Address {
			findedIndex = index
			isFinded = true
			break
		}
	}
	return findedIndex, isFinded
}

func (hc *HeathcheckEntity) GetServiceState(hcService *domain.HCService) (*domain.HCService, error) {
	hc.Lock()
	defer hc.Unlock()
	for _, runHC := range hc.runningHeathchecks {
		if runHC.Address == hcService.Address {
			return copyHCService(runHC), nil
		}
	}
	return nil, fmt.Errorf("get service state fail: service %v not found in healthchecks", hcService.Address)
}

func copyHCService(hcService *domain.HCService) *domain.HCService {
	copyOfApplicationServersFromService := getCopyOfApplicationServersFromService(hcService)
	return &domain.HCService{
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
		HCApplicationServers:  copyOfApplicationServersFromService,
		HCStop:                make(chan struct{}, 1),
		HCStopped:             make(chan struct{}, 1),
	}
}

func (hc *HeathcheckEntity) GetServicesState() ([]*domain.HCService, error) {
	hc.Lock()
	defer hc.Unlock()
	if len(hc.runningHeathchecks) == 0 {
		return nil, fmt.Errorf("active hc services: %v", len(hc.runningHeathchecks))
	}
	copyOfHCServices := make([]*domain.HCService, len(hc.runningHeathchecks))
	for i, runHC := range hc.runningHeathchecks {
		copyOfHCServices[i] = copyHCService(runHC)
	}
	return copyOfHCServices, nil
}
