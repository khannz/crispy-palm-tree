package usecase

// TODO: healthchecks != usecase!
// TODO: featute: check routes tunnles and syscfg exist
import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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
	StopAllHeatlthchecks chan struct{}
	AllHeatlthchecksDone chan struct{}
	runningHeathchecks   []*HCService
	cacheStorage         domain.StorageActions
	persistentStorage    domain.StorageActions
	ipvsadm              domain.IPVSWorker
	techInterface        *net.TCPAddr
	locker               *domain.Locker
	gracefulShutdown     *domain.GracefulShutdown
	dw                   *dummyWorker
	isMockMode           bool
	logging              *logrus.Logger
	announcedServices    map[string]int
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(cacheStorage domain.StorageActions,
	persistentStorage domain.StorageActions,
	ipvsadm domain.IPVSWorker,
	rawTechInterface string,
	locker *domain.Locker,
	gracefulShutdown *domain.GracefulShutdown,
	isMockMode bool,
	logging *logrus.Logger) *HeathcheckEntity {
	ti, _, _ := net.ParseCIDR(rawTechInterface + "/32")

	return &HeathcheckEntity{
		StopAllHeatlthchecks: make(chan struct{}, 1),
		AllHeatlthchecksDone: make(chan struct{}, 1),
		runningHeathchecks:   []*HCService{}, // need for append
		cacheStorage:         cacheStorage,
		persistentStorage:    persistentStorage,
		ipvsadm:              ipvsadm,
		techInterface:        &net.TCPAddr{IP: ti},
		locker:               locker,
		gracefulShutdown:     gracefulShutdown,
		dw:                   new(dummyWorker),
		isMockMode:           isMockMode,
		logging:              logging,
		announcedServices:    make(map[string]int),
	}
}

// UnknownDataStruct used for trying to get an unknown json or array of json's
type UnknownDataStruct struct {
	UnknowMap         map[string]string
	UnknowArrayOfMaps []map[string]string
}

func (hc *HeathcheckEntity) addNewServiceToMayAnnouncedServices(serviceIP string) {
	if _, inMap := hc.announcedServices[serviceIP]; inMap {
		return // already in map
	}
	hc.announcedServices[serviceIP] = 0 // add new service to annonced pool
}

// StartGracefulShutdownControlForHealthchecks ...
func (hc *HeathcheckEntity) StartGracefulShutdownControlForHealthchecks() {
	<-hc.StopAllHeatlthchecks
	hc.Lock()
	hc.gracefulShutdown.ShutdownNow = true
	hc.Unlock()
	for _, hcService := range hc.runningHeathchecks {
		hcService.HCStop <- struct{}{}
		hc.logging.Tracef("send stop checks for service %v", hcService.Address)
	}
	for _, hcService := range hc.runningHeathchecks {
		<-hcService.HCStopped
		hc.annonceLogic(hcService.IP, hcService.IsUp) // lock hc and dummy; may remove annonce
		hc.logging.Tracef("get checks stopped from service %v", hcService.Address)
	}
	hc.AllHeatlthchecksDone <- struct{}{}
}

// StartHealthchecksForCurrentServices ...
func (hc *HeathcheckEntity) StartHealthchecksForCurrentServices() error {
	hc.Lock()
	defer hc.Unlock()
	// if shutdown command at start
	if hc.gracefulShutdown.ShutdownNow {
		return nil
	}
	domainServicesInfo, err := hc.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return fmt.Errorf("fail when try LoadAllStorageDataToDomainModels: %v", err)
	}
	hcServices := make([]*HCService, len(domainServicesInfo))

	for i, domainServiceInfo := range domainServicesInfo {
		hcServices[i] = convertDomainServiceToHCService(domainServiceInfo)
	}

	for _, hcService := range hcServices {
		enrichApplicationServersHealthchecks(hcService, nil) // lock hcService
		hc.runningHeathchecks = append(hc.runningHeathchecks, hcService)
		hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
		go hc.startHealthchecksForCurrentService(hcService)
	}
	return nil
}

// NewServiceToHealtchecks - add service for healthchecks
func (hc *HeathcheckEntity) NewServiceToHealtchecks(domainService *domain.ServiceInfo) {
	hcService := convertDomainServiceToHCService(domainService)
	hc.Lock()
	defer hc.Unlock()
	if hc.gracefulShutdown.ShutdownNow {
		return
	}
	enrichApplicationServersHealthchecks(hcService, nil) // lock hcService
	hc.runningHeathchecks = append(hc.runningHeathchecks, hcService)
	hc.addNewServiceToMayAnnouncedServices(hcService.IP) // hc must be locked
	go hc.startHealthchecksForCurrentService(hcService)
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
func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(domainService *domain.ServiceInfo) {
	hcService := convertDomainServiceToHCService(domainService)
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
		hc.logging.Tracef("get checks stopped from service %v", hcService.Address)
		return
	}

	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": healthcheckID,
	}).Errorf("Heathcheck error: RemoveServiceFromHealtchecks error: service %v not found",
		hcService.Address)
}

// UpdateServiceAtHealtchecks ...
func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(domainService *domain.ServiceInfo) error {
	hcService := convertDomainServiceToHCService(domainService)
	hc.Lock()
	if hc.gracefulShutdown.ShutdownNow {
		hc.Unlock()
		return nil
	}
	updateIndex, isFinded := hc.findServiceInHealtcheckSlice(hcService.Address)
	hc.Unlock()
	if isFinded {
		hc.runningHeathchecks[updateIndex].Lock()
		hc.Lock()
		currentApplicationServers := getCopyOfApplicationServersFromService(hc.runningHeathchecks[updateIndex])
		hc.Unlock()
		hc.runningHeathchecks[updateIndex].Unlock()
		hc.logging.Tracef("send stop checks for service %v", hcService.Address)
		hc.runningHeathchecks[updateIndex].HCStop <- struct{}{}
		// TODO: annonce is ok?
		<-hc.runningHeathchecks[updateIndex].HCStopped
		hc.logging.Tracef("healthchecks stopped fr update from service %v", hcService.Address)
		enrichApplicationServersHealthchecks(hcService, currentApplicationServers) // lock hcService
		tmpHC := append(hc.runningHeathchecks[:updateIndex], hcService)
		hc.Lock()
		hc.runningHeathchecks = append(tmpHC, hc.runningHeathchecks[updateIndex+1:]...)
		hc.Unlock()
		go hc.startHealthchecksForCurrentService(hcService)
		return nil
	}
	return fmt.Errorf("Heathcheck error: UpdateServiceAtHealtchecks error: service %v not found",
		hcService.Address)
}

func getCopyOfApplicationServersFromService(serviceInfo *HCService) []*ApplicationServer {
	applicationServers := make([]*ApplicationServer, len(serviceInfo.ApplicationServers))
	for i, applicationServer := range serviceInfo.ApplicationServers {
		tmpHCApplicationServer := HCApplicationServer{
			HCType:           applicationServer.HCApplicationServer.HCType,
			HCAddress:        applicationServer.HCApplicationServer.HCAddress,
			HCTimeout:        applicationServer.HCApplicationServer.HCTimeout,
			RetriesForUP:     applicationServer.HCApplicationServer.RetriesForUP,
			RetriesForDown:   applicationServer.HCApplicationServer.RetriesForDown,
			LastIndexForUp:   applicationServer.HCApplicationServer.LastIndexForUp,
			LastIndexForDown: applicationServer.HCApplicationServer.LastIndexForDown,
			NearFieldsMode:   applicationServer.HCApplicationServer.NearFieldsMode,
			UserDefinedData:  applicationServer.HCApplicationServer.UserDefinedData,
		}

		applicationServers[i] = &ApplicationServer{
			Address:             applicationServer.Address,
			IP:                  applicationServer.IP,
			Port:                applicationServer.Port,
			IsUp:                applicationServer.IsUp,
			HCAddress:           applicationServer.HCAddress,
			HCApplicationServer: tmpHCApplicationServer,
			ExampleBashCommands: applicationServer.ExampleBashCommands,
		}
	}
	return applicationServers
}

func enrichApplicationServersHealthchecks(newServiceHealthcheck *HCService, oldApplicationServers []*ApplicationServer) {
	newServiceHealthcheck.Lock()
	defer newServiceHealthcheck.Unlock()
	newServiceHealthcheck.HCStop = make(chan struct{}, 1)
	newServiceHealthcheck.HCStopped = make(chan struct{}, 1)

	for i := range newServiceHealthcheck.ApplicationServers {
		hcApplicationServer := HCApplicationServer{}
		hcApplicationServer.HCType = newServiceHealthcheck.HCType
		hcApplicationServer.HCAddress = newServiceHealthcheck.ApplicationServers[i].HCAddress
		hcApplicationServer.HCTimeout = newServiceHealthcheck.HCTimeout
		hcApplicationServer.LastIndexForUp = 0
		hcApplicationServer.LastIndexForDown = 0
		hcApplicationServer.NearFieldsMode = newServiceHealthcheck.HCNearFieldsMode
		hcApplicationServer.UserDefinedData = newServiceHealthcheck.HCUserDefinedData

		retriesCounterForUp := make([]bool, newServiceHealthcheck.HCRetriesForUP)
		retriesCounterForDown := make([]bool, newServiceHealthcheck.HCRetriesForDown)

		j, isFinded := findApplicationServer(newServiceHealthcheck.Address, oldApplicationServers)
		if isFinded {
			fillNewBooleanArray(retriesCounterForUp, oldApplicationServers[j].HCApplicationServer.RetriesForUP)
			fillNewBooleanArray(retriesCounterForDown, oldApplicationServers[j].HCApplicationServer.RetriesForDown)
			hcApplicationServer.RetriesForUP = retriesCounterForUp
			hcApplicationServer.RetriesForDown = retriesCounterForDown
			newServiceHealthcheck.ApplicationServers[i].IsUp = oldApplicationServers[j].IsUp
			newServiceHealthcheck.ApplicationServers[i].HCApplicationServer = hcApplicationServer
			continue
		}
		hcApplicationServer.RetriesForUP = retriesCounterForUp
		hcApplicationServer.RetriesForDown = retriesCounterForDown
		newServiceHealthcheck.ApplicationServers[i].IsUp = false
		newServiceHealthcheck.ApplicationServers[i].HCApplicationServer = hcApplicationServer
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

func (hc *HeathcheckEntity) startHealthchecksForCurrentService(hcService *HCService) {
	// first run hc at create entity
	domainService := convertHCServiceToDomainServiceInfo(hcService) // TODO: that bad tmp solution
	hc.CheckApplicationServersInService(domainService)              // lock hc, hcService, dummy

	ticker := time.NewTicker(hcService.HCRepeat)
	for {
		select {
		case <-hcService.HCStop:
			hc.logging.Tracef("get stop checks command for service %v; send checks stoped and return", hcService.Address)
			hcService.HCStopped <- struct{}{}
			return
		case <-ticker.C:
			hc.CheckApplicationServersInService(domainService) // lock hc, hcService, dummy
		}
	}
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(domainService *domain.ServiceInfo) {
	hcService := convertDomainServiceToHCService(domainService)
	fs := &failedApplicationServers{wg: new(sync.WaitGroup)}

	for i := range hcService.ApplicationServers {
		fs.wg.Add(1)
		go hc.checkApplicationServerInService(hcService,
			fs,
			i) // lock hcService
	}
	fs.wg.Wait()
	percentageUp := percentageOfUp(len(hcService.ApplicationServers), fs.count)
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": healthcheckID,
	}).Debugf("Heathcheck: in service %v failed services is %v of %v; %v up percent of %v max for this service",
		hcService.Address,
		fs.count,
		len(hcService.ApplicationServers),
		percentageUp,
		hcService.AlivedAppServersForUp)
	isServiceUp := percentageOfDownBelowMPercentOfAlivedForUp(percentageUp, hcService.AlivedAppServersForUp)
	hc.logging.Debugf("Old service state %v. New service state %v", hcService.IsUp, isServiceUp)

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
		hc.logging.Debugf("service state not changed")
	}
	hc.updateInStorages(hcService)
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

func (hc *HeathcheckEntity) updateInStorages(hcService *HCService) {
	serviceInfo := convertHCServiceToDomainServiceInfo(hcService)
	errUpdataCache := hc.cacheStorage.UpdateServiceInfo(serviceInfo, healthcheckID)
	if errUpdataCache != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck update info in cache fail: %v", errUpdataCache)
	}

	errPersistantStorage := hc.persistentStorage.UpdateServiceInfo(serviceInfo, healthcheckID)
	if errPersistantStorage != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck update info in persistent storage fail: %v", errPersistantStorage)
	}
}

func (hc *HeathcheckEntity) checkApplicationServerInService(hcService *HCService,
	fs *failedApplicationServers,
	applicationServerInfoIndex int) {
	// TODO: to many code here.. Refactor to funcs
	defer fs.wg.Done()
	var isApplicationServerUp, isApplicationServerChangeState bool
	switch hcService.HCType {
	case "tcp":
		isCheckOk := hc.tcpCheckOk(hcService.ApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCTimeout)

		hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk) // lock hcService
		if !isCheckOk {
			isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
			hc.logging.Debugf("one hc for application server %v fail: %v; is change state: %v",
				hcService.ApplicationServers[applicationServerInfoIndex].Address,
				isApplicationServerUp,
				isApplicationServerChangeState)
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if isApplicationServerChangeState {
					if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.ApplicationServers[applicationServerInfoIndex]); err != nil {
						hc.logging.WithFields(logrus.Fields{
							"entity":   healthcheckName,
							"event id": healthcheckID,
						}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
					}
				}
			}
			return
		}

		isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
		hc.logging.Debugf("one hc for application server %v ok: %v; is change state: %v",
			hcService.ApplicationServers[applicationServerInfoIndex].Address,
			isApplicationServerUp,
			isApplicationServerChangeState)
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	case "http":
		isCheckOk := hc.httpCheckOk(hcService.ApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCTimeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
			isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if isApplicationServerChangeState {
					if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.ApplicationServers[applicationServerInfoIndex]); err != nil {
						hc.logging.WithFields(logrus.Fields{
							"entity":   healthcheckName,
							"event id": healthcheckID,
						}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
					}
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
		isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	case "http-advanced":
		isCheckOk := hc.httpAdvancedCheckOk(hcService.HCType,
			hcService.ApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCNearFieldsMode,
			hcService.HCUserDefinedData,
			hcService.HCTimeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
			isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if isApplicationServerChangeState {
					if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.ApplicationServers[applicationServerInfoIndex]); err != nil {
						hc.logging.WithFields(logrus.Fields{
							"entity":   healthcheckName,
							"event id": healthcheckID,
						}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
					}
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
		isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	case "icmp":
		isCheckOk := hc.icmpCheckOk(hcService.ApplicationServers[applicationServerInfoIndex].HCAddress,
			hcService.HCTimeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
			isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if isApplicationServerChangeState {
					if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.ApplicationServers[applicationServerInfoIndex]); err != nil {
						hc.logging.WithFields(logrus.Fields{
							"entity":   healthcheckName,
							"event id": healthcheckID,
						}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
					}
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoIndex, isCheckOk)                                                // lock hcService
		isApplicationServerUp, isApplicationServerChangeState = hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoIndex) // lock hcService
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck error: unknown healtcheck type: %v", hcService.HCType)
		return // must never will bfe. all data already validated
	}

	if isApplicationServerUp && isApplicationServerChangeState { // TODO: trace info TODO: do not UP when server already up!
		if err := hc.inclideApplicationServerInIPVS(hcService, hcService.ApplicationServers[applicationServerInfoIndex]); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": healthcheckID,
			}).Errorf("Heathcheck error: inclide application server in IPVS error: %v", err)
		}
		return
	}
}

func (hc *HeathcheckEntity) isApplicationServerUpAndStateChange(hcService *HCService,
	applicationServerInfoIndex int) (bool, bool) {
	//return isUp and isChagedState booleans
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("real: %v, RetriesCounterForDown: %v", hcService.ApplicationServers[applicationServerInfoIndex].Address, hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForUP)
	hc.logging.Tracef("real: %v, RetriesCounterForUp: %v", hcService.ApplicationServers[applicationServerInfoIndex].Address, hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForDown)

	if hcService.ApplicationServers[applicationServerInfoIndex].IsUp {
		// check it not down
		for _, isUp := range hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForDown {
			if isUp {
				return true, false // do not change up state
			}
		}
		hcService.ApplicationServers[applicationServerInfoIndex].IsUp = false // if all hc fail at RetriesCounterForDown - change state
		hc.logging.WithFields(logrus.Fields{
			"event id": healthcheckID,
		}).Warnf("at service %v real server %v DOWN", hcService.Address,
			hcService.ApplicationServers[applicationServerInfoIndex].Address)
		return false, true
	}

	for _, isUp := range hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForUP {
		if !isUp {
			// do not change down state
			return false, false
		}
	}

	// all RetriesCounterForUp true
	hcService.ApplicationServers[applicationServerInfoIndex].IsUp = true // if all hc fail at RetriesCounterForDown - change state
	hc.logging.WithFields(logrus.Fields{
		"event id": healthcheckID,
	}).Warnf("at service %v real server %v UP", hcService.Address,
		hcService.ApplicationServers[applicationServerInfoIndex].Address)
	return true, true

}

func (hc *HeathcheckEntity) moveApplicationServerStateIndexes(hcService *HCService, applicationServerInfoIndex int, isUpNow bool) {
	hcService.Lock()
	defer hcService.Unlock()
	if len(hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForUP) < hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForUp+1 {
		hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForUp = 0
	}
	hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForUP[hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForUp] = isUpNow
	hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForUp++

	if len(hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForDown) < hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForDown+1 {
		hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForDown = 0
	}
	hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.RetriesForDown[hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForDown] = isUpNow
	hcService.ApplicationServers[applicationServerInfoIndex].HCApplicationServer.LastIndexForDown++
}

func (hc *HeathcheckEntity) tcpCheckOk(healthcheckAddress string, timeout time.Duration) bool {
	hcSlice := strings.Split(healthcheckAddress, ":")
	hcPort := ""
	if len(hcSlice) > 1 {
		hcPort = hcSlice[1]
	}
	hcIP := hcSlice[0]
	dialer := net.Dialer{
		LocalAddr: hc.techInterface,
		Timeout:   timeout}

	conn, err := dialer.Dial("tcp", net.JoinHostPort(hcIP, hcPort))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: Connecting tcp connect error: %v", err)
		return false
	}
	defer conn.Close()

	if conn != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck info port opened: %v", net.JoinHostPort(hcIP, hcPort))
		return true
	}

	// somehow it can be..
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": healthcheckID,
	}).Error("Heathcheck has unknown error: connection is nil, but have no errors")
	return false
}

func (hc *HeathcheckEntity) httpCheckOk(healthcheckAddress string, timeout time.Duration) bool {
	// FIXME: https checks also here
	roundTripper := &http.Transport{
		Dial: (&net.Dialer{
			LocalAddr: hc.techInterface,
			Timeout:   timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout * 6 / 10, // hardcode
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: roundTripper,
		Timeout:   timeout,
	}
	resp, err := client.Get(healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: Connecting http error: %v", err)
		return false
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return false
	}
	return true
}

func (hc *HeathcheckEntity) icmpCheckOk(healthcheckAddress string, timeout time.Duration) bool {
	// Start listening for icmp replies
	icpmConnection, err := icmp.ListenPacket("ip4:icmp", hc.techInterface.String())
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm connection error: %v", err)
		return false
	}
	defer icpmConnection.Close()

	// Get the real IP of the target
	dst, err := net.ResolveIPAddr("ip4", healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm resolve ip addr error: %v", err)
		return false
	}

	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  1,
			Data: []byte("hello")},
	}

	b, err := m.Marshal(nil)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm marshall message error: %v", err)
		return false
	}

	// Send it
	n, err := icpmConnection.WriteTo(b, dst)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm write bytes to error: %v", err)
		return false
	} else if n != len(b) {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm write bytes to error (not all of bytes was send): %v", err)
		return false
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = icpmConnection.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm set read deadline error: %v", err)
		return false
	}
	n, peer, err := icpmConnection.ReadFrom(reply)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm read reply error: %v", err)
		return false
	}

	// Let's look what we have in reply
	rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm parse message error: %v", err)
		return false
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck icpm for %v succes", healthcheckAddress)
		return true
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: icpm for %v reply type error: got %+v from %v; want echo reply",
			healthcheckAddress,
			rm,
			peer)
		return false
	}
}

func (hc *HeathcheckEntity) excludeApplicationServerFromIPVS(hcService *HCService,
	applicationServer *ApplicationServer) error {
	tmpAppSs := convertHCApplicationServersToDomainApplicationServers([]*ApplicationServer{applicationServer})
	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(hcService.IP,
		hcService.IP,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol,
		tmpAppSs)
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

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(hcService *HCService,
	applicationServer *ApplicationServer) error {
	tmpAppSs := convertHCApplicationServersToDomainApplicationServers([]*ApplicationServer{applicationServer})
	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(hcService.IP,
		hcService.Port,
		hcService.RoutingType,
		hcService.BalanceType,
		hcService.Protocol,
		tmpAppSs)
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err = hc.ipvsadm.AddApplicationServersForService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		healthcheckID); err != nil {
		return fmt.Errorf("Error when ipvsadm add application servers for service: %v", err)
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

// http advanced start
func (hc *HeathcheckEntity) httpAdvancedCheckOk(hcType string,
	hcAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration) bool {
	switch hcType {
	case "http-advanced-json":
		return hc.httpAdvancedJSONCheckOk(hcAddress, nearFieldsMode,
			userDefinedData, timeout)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck error: http advanced check fail error: unknown check type: %v", hcType)
		return false
	}
}

func (hc *HeathcheckEntity) IsMockMode() bool {
	return hc.isMockMode
}

// http advanced json start
func (hc *HeathcheckEntity) httpAdvancedJSONCheckOk(hcAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", hcAddress, nil)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Errorf("Heathcheck error: http advanced JSON check fail error: can't make new http request: %v", err)
		return false
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: Connecting http advanced JSON check error: %v", err)
		return false
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return false
	}

	u := UnknownDataStruct{}
	if err := json.Unmarshal(response, &u.UnknowMap); err != nil {
		if err := json.Unmarshal(response, &u.UnknowArrayOfMaps); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": healthcheckID,
			}).Tracef("Heathcheck error: http advanced JSON check fail error: can't unmarshal response from: %v, error: %v",
				hcAddress,
				err)
			return false
		}
	}

	if u.UnknowMap == nil && u.UnknowArrayOfMaps == nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck error: http advanced JSON check fail error: response is nil from: %v", hcAddress)
		return false
	}

	if nearFieldsMode { // mode for finding all matches for the desired object in a single map
		if hc.isFinderForNearFieldsModeFail(userDefinedData, u, hcAddress) { // if false do not return, continue range params
			return false
		}
	} else {
		if hc.isFinderMapToMapFail(userDefinedData, u, hcAddress) { // if false do not return, continue range params
			return false
		}
	}

	return true
}

func (hc *HeathcheckEntity) isFinderForNearFieldsModeFail(userSearchData map[string]string,
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string) bool {
	numberOfRequiredMatches := len(userSearchData) // the number of required matches in the user's search map
	var mapForSearch map[string]string             // the map that we will use to search for all matches(beacose that nearFieldsMode)
	for sK, sV := range userSearchData {           // go through the search map
		if numberOfRequiredMatches != 0 { // checking that not all matches are found within the search map
			if mapForSearch != nil {
				if isKVequal(sK, sV, mapForSearch) {
					numberOfRequiredMatches-- // reduced search by length of matches
				}
			} else { // If matches haven't been found yet (nearFieldsMode)
				if unknownDataStruct.UnknowArrayOfMaps != nil {
					for _, incomeData := range unknownDataStruct.UnknowArrayOfMaps { // go through the array of maps received on request
						if isKVequal(sK, sV, incomeData) {
							numberOfRequiredMatches-- // reduced search by length of matches
							mapForSearch = incomeData // other matches for the desired map will be searched only in this one (nearFieldsMode)
							break
						}
					}
				} else if unknownDataStruct.UnknowMap != nil {
					if isKVequal(sK, sV, unknownDataStruct.UnknowMap) {
						numberOfRequiredMatches--                  // reduced search by length of matches
						mapForSearch = unknownDataStruct.UnknowMap // other matches for the desired map will be searched only in this one (nearFieldsMode)
					}
				}
			}
		}
	}
	if numberOfRequiredMatches != 0 {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)
		return true
	}
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": healthcheckID,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func (hc *HeathcheckEntity) isFinderMapToMapFail(userSearchData map[string]string,
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string) bool {
	numberOfRequiredMatches := len(userSearchData) // the number of required matches in the user's search map

	for sK, sV := range userSearchData { // go through the search map
		if numberOfRequiredMatches != 0 { // checking that not all matches are found within the search map
			if unknownDataStruct.UnknowArrayOfMaps != nil {
				for _, incomeData := range unknownDataStruct.UnknowArrayOfMaps { // go through the array of maps received on request
					if isKVequal(sK, sV, incomeData) {
						numberOfRequiredMatches-- // reduced search by length of matches
						break
					}
				}
			} else if unknownDataStruct.UnknowMap != nil {
				if isKVequal(sK, sV, unknownDataStruct.UnknowMap) {
					numberOfRequiredMatches-- // reduced search by length of matches
				}
			}
		}
	}

	if numberOfRequiredMatches != 0 {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": healthcheckID,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)

		return true
	}
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": healthcheckID,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func isKVequal(k string, v interface{}, mapForSearch map[string]string) bool {
	if mI, isKeyFinded := mapForSearch[k]; isKeyFinded {
		if v == mI {
			return true
		}
	}
	return false
}

// http advanced json end

// http advanced end

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

func findApplicationServer(address string, oldApplicationServers []*ApplicationServer) (int, bool) {
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
