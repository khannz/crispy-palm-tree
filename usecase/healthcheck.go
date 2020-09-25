package usecase

// TODO: healthchecks != usecase!
// TODO: featute: check routes tunnles and syscfg exist
import (
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
const healthcheckUUID = "00000000-0000-0000-0000-000000000004"
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
	runningHeathchecks []*domain.ServiceInfo
	cacheStorage       domain.StorageActions
	persistentStorage  domain.StorageActions
	ipvsadm            domain.IPVSWorker
	techInterface      *net.TCPAddr
	locker             *domain.Locker
	gracefulShutdown   *domain.GracefulShutdown
	dw                 *dummyWorker
	isMockMode         bool
	logging            *logrus.Logger
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
		runningHeathchecks: []*domain.ServiceInfo{}, // need for append
		cacheStorage:       cacheStorage,
		persistentStorage:  persistentStorage,
		ipvsadm:            ipvsadm,
		techInterface:      &net.TCPAddr{IP: ti},
		locker:             locker,
		gracefulShutdown:   gracefulShutdown,
		dw:                 new(dummyWorker),
		isMockMode:         isMockMode,
		logging:            logging,
	}
}

// UnknownDataStruct used for trying to get an unknown json or array of json's
type UnknownDataStruct struct {
	UnknowMap         map[string]interface{}
	UnknowArrayOfMaps []map[string]interface{}
}

// StartGracefulShutdownControlForHealthchecks ...
func (hc *HeathcheckEntity) StartGracefulShutdownControlForHealthchecks() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		hc.Lock()
		if hc.gracefulShutdown.ShutdownNow {
			for _, serviceInfo := range hc.runningHeathchecks {
				serviceInfo.Lock()
				serviceInfo.Healthcheck.StopChecks <- struct{}{}
				serviceInfo.Unlock()
			}
		}
		hc.Unlock()
	}
}

// StartHealthchecksForCurrentServices ...
func (hc *HeathcheckEntity) StartHealthchecksForCurrentServices() error {
	hc.Lock()
	defer hc.Unlock()
	// if shutdown command at start
	if hc.gracefulShutdown.ShutdownNow {
		return nil
	}
	servicesInfo, err := hc.cacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		return fmt.Errorf("fail when try LoadAllStorageDataToDomainModels: %v", err)
	}

	for _, serviceInfo := range servicesInfo {
		enrichApplicationServersHealthchecks(serviceInfo, nil) // locks inside
		hc.runningHeathchecks = append(hc.runningHeathchecks, serviceInfo)
		go hc.startHealthchecksForCurrentService(serviceInfo)
	}
	return nil
}

// NewServiceToHealtchecks - add service for healthchecks
func (hc *HeathcheckEntity) NewServiceToHealtchecks(serviceInfo *domain.ServiceInfo) {
	hc.Lock()
	defer hc.Unlock()
	if hc.gracefulShutdown.ShutdownNow {
		return
	}
	enrichApplicationServersHealthchecks(serviceInfo, nil) // locks inside
	hc.runningHeathchecks = append(hc.runningHeathchecks, serviceInfo)
	go hc.startHealthchecksForCurrentService(serviceInfo)
}

// RemoveServiceFromHealtchecks ...
func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(serviceInfo *domain.ServiceInfo) {
	hc.Lock()
	defer hc.Unlock()
	if hc.gracefulShutdown.ShutdownNow {
		return
	}
	indexForRemove, isFinded := hc.findServiceInHealtcheckSlice(serviceInfo.ServiceIP, serviceInfo.ServicePort)
	if isFinded {
		hc.runningHeathchecks[indexForRemove].Lock()
		hc.runningHeathchecks[indexForRemove].Healthcheck.StopChecks <- struct{}{}
		hc.runningHeathchecks[indexForRemove].Unlock()
		hc.runningHeathchecks = append(hc.runningHeathchecks[:indexForRemove], hc.runningHeathchecks[indexForRemove+1:]...)
	} else {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: RemoveServiceFromHealtchecks error: service %v:%v not found",
			serviceInfo.ServiceIP,
			serviceInfo.ServicePort)
	}
}

// UpdateServiceAtHealtchecks ...
func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(serviceInfo *domain.ServiceInfo) error {
	hc.Lock()
	defer hc.Unlock()
	if hc.gracefulShutdown.ShutdownNow {
		return nil
	}
	updateIndex, isFinded := hc.findServiceInHealtcheckSlice(serviceInfo.ServiceIP, serviceInfo.ServicePort)
	if isFinded {
		hc.runningHeathchecks[updateIndex].Lock()
		currentApplicationServers := getCopyOfApplicationServersFromService(hc.runningHeathchecks[updateIndex])
		hc.runningHeathchecks[updateIndex].Healthcheck.StopChecks <- struct{}{}
		hc.runningHeathchecks[updateIndex].Unlock()
		enrichApplicationServersHealthchecks(serviceInfo, currentApplicationServers) // locks inside
		tmpHC := append(hc.runningHeathchecks[:updateIndex], serviceInfo)
		hc.runningHeathchecks = append(tmpHC, hc.runningHeathchecks[updateIndex+1:]...)
		go hc.startHealthchecksForCurrentService(serviceInfo)
		return nil
	}
	return fmt.Errorf("Heathcheck error: UpdateServiceAtHealtchecks error: service %v:%v not found",
		serviceInfo.ServiceIP,
		serviceInfo.ServicePort)
}

func getCopyOfApplicationServersFromService(serviceInfo *domain.ServiceInfo) []domain.ApplicationServer {
	applicationServers := make([]domain.ApplicationServer, len(serviceInfo.ApplicationServers))
	for i, applicationServer := range serviceInfo.ApplicationServers {
		applicationServers[i] = *applicationServer
	}
	return applicationServers
}

func enrichApplicationServersHealthchecks(newServiceHealthcheck *domain.ServiceInfo, oldApplicationServers []domain.ApplicationServer) {
	newServiceHealthcheck.Lock()
	defer newServiceHealthcheck.Unlock()
	newServiceHealthcheck.Healthcheck.StopChecks = make(chan struct{}, 1)
	for i := range newServiceHealthcheck.ApplicationServers {
		retriesCounterForUp := make([]bool, newServiceHealthcheck.Healthcheck.RetriesForUpApplicationServer)
		retriesCounterForDown := make([]bool, newServiceHealthcheck.Healthcheck.RetriesForDownApplicationServer)
		lastIndexForUp := 0
		lastIndexForDown := 0
		newServiceHealthcheck.ApplicationServers[i].ServerHealthcheck.LastIndexForUp = lastIndexForUp // FIXME: invalid memory address or nil pointer dereference
		newServiceHealthcheck.ApplicationServers[i].ServerHealthcheck.LastIndexForDown = lastIndexForDown

		j, isFinded := findApplicationServer(newServiceHealthcheck.ServiceIP, newServiceHealthcheck.ServicePort, oldApplicationServers)
		if isFinded {
			fillNewBooleanArray(retriesCounterForUp, oldApplicationServers[j].ServerHealthcheck.RetriesCounterForUp)
			fillNewBooleanArray(retriesCounterForDown, oldApplicationServers[j].ServerHealthcheck.RetriesCounterForDown)
			newServiceHealthcheck.ApplicationServers[i].ServerHealthcheck.RetriesCounterForUp = retriesCounterForUp
			newServiceHealthcheck.ApplicationServers[i].ServerHealthcheck.RetriesCounterForDown = retriesCounterForDown
			newServiceHealthcheck.ApplicationServers[i].IsUp = oldApplicationServers[j].IsUp
			continue
		}
		newServiceHealthcheck.ApplicationServers[i].ServerHealthcheck.RetriesCounterForUp = retriesCounterForUp
		newServiceHealthcheck.ApplicationServers[i].ServerHealthcheck.RetriesCounterForDown = retriesCounterForDown
		newServiceHealthcheck.ApplicationServers[i].IsUp = false
	}
}

func (hc *HeathcheckEntity) findServiceInHealtcheckSlice(serviceIP, servicePort string) (int, bool) {
	var findedIndex int
	var isFinded bool
	for index, runningServiceHc := range hc.runningHeathchecks {
		if serviceIP == runningServiceHc.ServiceIP &&
			servicePort == runningServiceHc.ServicePort {
			findedIndex = index
			isFinded = true
			break
		}
	}
	return findedIndex, isFinded
}

func (hc *HeathcheckEntity) startHealthchecksForCurrentService(serviceInfo *domain.ServiceInfo) {
	hc.gracefulShutdown.Lock()
	hc.gracefulShutdown.UsecasesJobs++
	hc.gracefulShutdown.Unlock()
	defer decreaseJobs(hc.gracefulShutdown)

	// first run hc at create entity
	hc.CheckApplicationServersInService(serviceInfo) // locks inside

	ticker := time.NewTicker(serviceInfo.Healthcheck.RepeatHealthcheck)
	for {
		select {
		case <-serviceInfo.Healthcheck.StopChecks:
			return
		case <-ticker.C:
			hc.CheckApplicationServersInService(serviceInfo)
		}
	}
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(serviceInfo *domain.ServiceInfo) {
	fs := &failedApplicationServers{wg: new(sync.WaitGroup)}
	for i, applicationServerInfo := range serviceInfo.ApplicationServers {
		fs.wg.Add(1)
		go hc.checkApplicationServerInService(serviceInfo,
			applicationServerInfo,
			fs,
			i) // locks inside
	}
	fs.wg.Wait()
	percentageDown := percentageOfDown(len(serviceInfo.ApplicationServers), fs.count)
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Debugf("Heathcheck: in service %v:%v failed services is %v of %v; %v failed percent of %v max for this service",
		serviceInfo.ServiceIP,
		serviceInfo.ServicePort,
		fs.count,
		len(serviceInfo.ApplicationServers),
		percentageDown,
		serviceInfo.Healthcheck.PercentOfAlivedForUp)
	if percentageDown > serviceInfo.Healthcheck.PercentOfAlivedForUp {
		serviceInfo.IsUp = false
		hc.removeFromDummyWrapper(serviceInfo.ServiceIP) // FIXME: do not touch dummy
	} else {
		serviceInfo.IsUp = true
		hc.addToDummyWrapper(serviceInfo.ServiceIP) // FIXME: do not touch dummy
	}
	hc.updateInStorages(serviceInfo)
}

func (hc *HeathcheckEntity) checkApplicationServerInService(serviceInfo *domain.ServiceInfo,
	applicationServerInfo *domain.ApplicationServer,
	fs *failedApplicationServers,
	applicationServerInfoIndex int) {
	// TODO: to many code here.. Refactor to funcs
	defer fs.wg.Done()
	var isApplicationServerUp bool
	switch serviceInfo.Healthcheck.Type {
	case "tcp":
		isCheckOk := hc.tcpCheckOk(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, !isCheckOk) // locks inside
			isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
		isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	case "http":
		isCheckOk := hc.httpCheckOk(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
			isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
		isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	case "http-advanced":
		isCheckOk := hc.httpAdvancedCheckOk(applicationServerInfo.ServerHealthcheck,
			serviceInfo.Healthcheck.Timeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
			isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
		isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	case "icmp":
		isCheckOk := hc.icmpCheckOk(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout)
		if !isCheckOk {
			hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
			isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
			if !isApplicationServerUp {
				fs.Lock()
				fs.count++
				fs.Unlock()
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
			return
		}
		hc.moveApplicationServerStateIndexes(serviceInfo, applicationServerInfoIndex, isCheckOk) // locks inside
		isApplicationServerUp = hc.isApplicationServerUp(serviceInfo, applicationServerInfoIndex)
		if !isApplicationServerUp {
			fs.Lock()
			fs.count++
			fs.Unlock()
		}
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: unknown healtcheck type: %v", serviceInfo.Healthcheck.Type)
		return // must never will be. all data already validated
	}

	if isApplicationServerUp { // TODO: trace info TODO: do not UP when server already up!
		if err := hc.inclideApplicationServerInIPVS(serviceInfo, applicationServerInfo); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Errorf("Heathcheck error: inclide application server in IPVS error: %v", err)
		}
		return
	}
}

func (hc *HeathcheckEntity) isApplicationServerUp(serviceInfo *domain.ServiceInfo,
	applicationServerInfoIndex int) bool {
	serviceInfo.Lock()
	defer serviceInfo.Unlock()
	if serviceInfo.ApplicationServers[applicationServerInfoIndex].IsUp {
		// check it not down
		for _, state := range serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.RetriesCounterForDown {
			if state { // at least one hc ok
				return true // do not change up state
			}
		}
		serviceInfo.ApplicationServers[applicationServerInfoIndex].IsUp = false // if all hc fail at RetriesCounterForDown - change state
		hc.logging.WithFields(logrus.Fields{
			"event uuid": healthcheckUUID,
		}).Warnf("at service %v:%v real server %v:%v DOWN", serviceInfo.ServiceIP,
			serviceInfo.ServicePort,
			serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerIP,
			serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerPort)
		return false
	}

	for _, state := range serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.RetriesCounterForUp {
		if state { // at least one hc ok
			serviceInfo.ApplicationServers[applicationServerInfoIndex].IsUp = true // if all hc fail at RetriesCounterForDown - change state
			hc.logging.WithFields(logrus.Fields{
				"event uuid": healthcheckUUID,
			}).Warnf("at service %v:%v real server %v:%v UP", serviceInfo.ServiceIP,
				serviceInfo.ServicePort,
				serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerIP,
				serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerPort)
			return true
		}
	}
	return false
	// do not change down state
}

func (hc *HeathcheckEntity) moveApplicationServerStateIndexes(serviceInfo *domain.ServiceInfo, applicationServerInfoIndex int, isUpNow bool) {
	serviceInfo.Lock()
	defer serviceInfo.Unlock()
	if len(serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.RetriesCounterForUp) < serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForUp+1 {
		serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForUp = 0
	}
	serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.RetriesCounterForUp[serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForUp] = isUpNow
	serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForUp++

	if len(serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.RetriesCounterForDown) < serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForDown+1 {
		serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForDown = 0
	}
	serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.RetriesCounterForDown[serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForDown] = isUpNow
	serviceInfo.ApplicationServers[applicationServerInfoIndex].ServerHealthcheck.LastIndexForDown++
}

func (hc *HeathcheckEntity) updateInStorages(serviceInfo *domain.ServiceInfo) {
	errUpdataCache := hc.cacheStorage.UpdateServiceInfo(serviceInfo, healthcheckUUID)
	if errUpdataCache != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck update info in cache fail: %v", errUpdataCache)
	}

	errPersistantStorage := hc.persistentStorage.UpdateServiceInfo(serviceInfo, healthcheckUUID)
	if errPersistantStorage != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck update info in persistent storage fail: %v", errPersistantStorage)
	}
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
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Connecting tcp connect error: %v", err)
		return false
	}
	defer conn.Close()

	if conn != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck info port opened: %v", net.JoinHostPort(hcIP, hcPort))
		return true
	}

	// somehow it can be..
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Error("Heathcheck has unknown error: connection is nil, but have no errors")
	return false
}

func (hc *HeathcheckEntity) httpCheckOk(healthcheckAddress string, timeout time.Duration) bool {
	// FIXME:  dialer := net.Dialer{
	// 	LocalAddr: hc.techInterface,
	// 	Timeout:   timeout}
	client := http.Client{
		// Transport: dialer,
		Timeout: timeout,
	}
	resp, err := client.Get(healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Connecting http error: %v", err)
		return false
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
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
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm connection error: %v", err)
		return false
	}
	defer icpmConnection.Close()

	// Get the real IP of the target
	dst, err := net.ResolveIPAddr("ip4", healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
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
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm marshall message error: %v", err)
		return false
	}

	// Send it
	n, err := icpmConnection.WriteTo(b, dst)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm write bytes to error: %v", err)
		return false
	} else if n != len(b) {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm write bytes to error (not all of bytes was send): %v", err)
		return false
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = icpmConnection.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm set read deadline error: %v", err)
		return false
	}
	n, peer, err := icpmConnection.ReadFrom(reply)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm read reply error: %v", err)
		return false
	}

	// Let's look what we have in reply
	rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm parse message error: %v", err)
		return false
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck icpm for %v succes", healthcheckAddress)
		return true
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm for %v reply type error: got %+v from %v; want echo reply",
			healthcheckAddress,
			rm,
			peer)
		return false
	}
}

func (hc *HeathcheckEntity) excludeApplicationServerFromIPVS(serviceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) error {
	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(serviceInfo.ServiceIP,
		serviceInfo.ServicePort,
		serviceInfo.RoutingType,
		serviceInfo.BalanceType,
		serviceInfo.Protocol,
		[]*domain.ApplicationServer{applicationServer})
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err := hc.ipvsadm.RemoveApplicationServersFromService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		healthcheckUUID); err != nil {
		return fmt.Errorf("Error when ipvsadm remove application servers from service: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(allServiceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) error {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	vip, port, routingType, balanceType, protocol, applicationServers, err := domain.PrepareDataForIPVS(formedServiceData.ServiceIP,
		formedServiceData.ServicePort,
		formedServiceData.RoutingType,
		formedServiceData.BalanceType,
		formedServiceData.Protocol,
		formedServiceData.ApplicationServers)
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err = hc.ipvsadm.AddApplicationServersForService(vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		healthcheckUUID); err != nil {
		return fmt.Errorf("Error when ipvsadm add application servers for service: %v", err)
	}
	return nil

}

func percentageOfDown(total, down int) int {
	if total == down {
		return 100
	}
	if down == 0 {
		return 0
	}
	return (total - down) * 100 / total
}

// http advanced start
func (hc *HeathcheckEntity) httpAdvancedCheckOk(serverHealthcheck domain.ServerHealthcheck,
	timeout time.Duration) bool {
	switch serverHealthcheck.TypeOfCheck {
	case "http-advanced-json":
		return hc.httpAdvancedJSONCheckOk(serverHealthcheck, timeout)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: http advanced check fail error: unknown check type: %v", serverHealthcheck.TypeOfCheck)
		return false
	}
}

func (hc *HeathcheckEntity) IsMockMode() bool {
	return hc.isMockMode
}

// http advanced json start
func (hc *HeathcheckEntity) httpAdvancedJSONCheckOk(serverHealthcheck domain.ServerHealthcheck,
	timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", serverHealthcheck.HealthcheckAddress, nil)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: http advanced JSON check fail error: can't make new http request: %v", err)
		return false
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Connecting http advanced JSON check error: %v", err)
		return false
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return false
	}

	u := UnknownDataStruct{}
	if err := json.Unmarshal(response, &u.UnknowMap); err != nil {
		if err := json.Unmarshal(response, &u.UnknowArrayOfMaps); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Tracef("Heathcheck error: http advanced JSON check fail error: can't unmarshal response from: %v, error: %v",
				serverHealthcheck.HealthcheckAddress,
				err)
			return false
		}
	}

	if u.UnknowMap == nil && u.UnknowArrayOfMaps == nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: http advanced JSON check fail error: response is nil from: %v", serverHealthcheck.HealthcheckAddress)
		return false
	}

	for _, aHP := range serverHealthcheck.AdvancedHealthcheckParameters { // go through the array of search objects
		if aHP.NearFieldsMode { // mode for finding all matches for the desired object in a single map
			if hc.isFinderForNearFieldsModeFail(aHP.UserDefinedData, u, serverHealthcheck.HealthcheckAddress) { // if false do not return, continue range params
				return false
			}
		} else {
			if hc.isFinderMapToMapFail(aHP.UserDefinedData, u, serverHealthcheck.HealthcheckAddress) { // if false do not return, continue range params
				return false
			}
		}
	}
	return true
}

func (hc *HeathcheckEntity) isFinderForNearFieldsModeFail(userSearchData map[string]interface{},
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string) bool {
	numberOfRequiredMatches := len(userSearchData) // the number of required matches in the user's search map
	var mapForSearch map[string]interface{}        // the map that we will use to search for all matches(beacose that nearFieldsMode)
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
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)
		return true
	}
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func (hc *HeathcheckEntity) isFinderMapToMapFail(userSearchData map[string]interface{},
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
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)

		return true
	}
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func isKVequal(k string, v interface{}, mapForSearch map[string]interface{}) bool {
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

func findApplicationServer(serverIP, serverPort string, oldApplicationServers []domain.ApplicationServer) (int, bool) {
	var findedIndex int
	var isFinded bool
	if oldApplicationServers == nil {
		return findedIndex, isFinded
	}
	for index, oldApplicationServer := range oldApplicationServers {
		if serverIP == oldApplicationServer.ServerIP &&
			serverPort == oldApplicationServer.ServerPort {
			findedIndex = index
			isFinded = true
			break
		}
	}
	return findedIndex, isFinded
}
