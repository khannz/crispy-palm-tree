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

type failedApplicationServers struct {
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
		serviceInfo.Lock()
		serviceInfo.Healthcheck.StopChecks = make(chan struct{}, 1)
		serviceInfo.Unlock()
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
	serviceInfo.Lock()
	serviceInfo.Healthcheck.StopChecks = make(chan struct{}, 1)
	serviceInfo.Unlock()
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
		hc.runningHeathchecks[updateIndex].Healthcheck.StopChecks <- struct{}{}
		hc.runningHeathchecks[updateIndex].Unlock()
		serviceInfo.Lock()
		serviceInfo.Healthcheck.StopChecks = make(chan struct{}, 1)
		serviceInfo.Unlock()
		tmpHC := append(hc.runningHeathchecks[:updateIndex], serviceInfo)
		hc.runningHeathchecks = append(tmpHC, hc.runningHeathchecks[updateIndex+1:]...)
		go hc.startHealthchecksForCurrentService(serviceInfo)
		return nil
	}
	return fmt.Errorf("Heathcheck error: UpdateServiceAtHealtchecks error: service %v:%v not found",
		serviceInfo.ServiceIP,
		serviceInfo.ServicePort)
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
	serviceInfo.Lock()
	hc.CheckApplicationServersInService(serviceInfo)
	serviceInfo.Unlock()

	ticker := time.NewTicker(serviceInfo.Healthcheck.RepeatHealthcheck)
	for {
		select {
		case <-serviceInfo.Healthcheck.StopChecks:
			return
		case <-ticker.C:
			serviceInfo.Lock()
			hc.CheckApplicationServersInService(serviceInfo)
			serviceInfo.Unlock()
		}
	}
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(serviceInfo *domain.ServiceInfo) {
	fs := &failedApplicationServers{wg: new(sync.WaitGroup)}
	for _, applicationServerInfo := range serviceInfo.ApplicationServers {
		fs.wg.Add(1)
		go hc.checkApplicationServerInService(serviceInfo,
			applicationServerInfo,
			fs)
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
		hc.removeFromDummyWrapper(serviceInfo.ServiceIP)
	} else {
		serviceInfo.IsUp = true
		hc.addToDummyWrapper(serviceInfo.ServiceIP)
	}
	hc.updateInStorages(serviceInfo)
}

func (hc *HeathcheckEntity) checkApplicationServerInService(serviceInfo *domain.ServiceInfo,
	applicationServerInfo *domain.ApplicationServer,
	fs *failedApplicationServers) {
	defer fs.wg.Done()
	switch serviceInfo.Healthcheck.Type {
	case "tcp":
		if hc.tcpCheckFail(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout) {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
			}
			if applicationServerInfo.IsUp {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down, healthcheck to %v fail",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
					applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
				applicationServerInfo.IsUp = false
				// if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				// 	hc.logging.WithFields(logrus.Fields{
				// 		"entity":     healthcheckName,
				// 		"event uuid": healthcheckUUID,
				// 	}).Debugf("Heathcheck error: excludeApplicationServerFromIPVS error: %v", err)
				// }
			}
			return
		}
	case "http":
		if hc.httpCheckFail(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout) {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
			}

			if applicationServerInfo.IsUp {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down, healthcheck to %v fail",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
					applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
				applicationServerInfo.IsUp = false
				// if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				// 	hc.logging.WithFields(logrus.Fields{
				// 		"entity":     healthcheckName,
				// 		"event uuid": healthcheckUUID,
				// 	}).Debugf("Heathcheck error: excludeApplicationServerFromIPVS error: %v", err)
				// }
			}
			return
		}
	case "http-advanced":
		if hc.httpAdvancedCheckFail(applicationServerInfo.ServerHealthcheck,
			serviceInfo.Healthcheck.Timeout) {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
			}
			if applicationServerInfo.IsUp {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down, healthcheck to %v fail",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
					applicationServerInfo.ServerHealthcheck)
				applicationServerInfo.IsUp = false
				// if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				// 	hc.logging.WithFields(logrus.Fields{
				// 		"entity":     healthcheckName,
				// 		"event uuid": healthcheckUUID,
				// 	}).Debugf("Heathcheck error: excludeApplicationServerFromIPVS error: %v", err)
				// }
			}
			return
		}
	case "icmp":
		if hc.icmpCheckFail(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout) {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
			}
			if applicationServerInfo.IsUp {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down, healthcheck to %v fail",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
					applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
				applicationServerInfo.IsUp = false
				// if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				// 	hc.logging.WithFields(logrus.Fields{
				// 		"entity":     healthcheckName,
				// 		"event uuid": healthcheckUUID,
				// 	}).Debugf("Heathcheck error: excludeApplicationServerFromIPVS error: %v", err)
				// }
			}
			return
		}
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: unknown healtcheck type: %v", serviceInfo.Healthcheck.Type)
		return // must never will be. all data already validated
	}

	if !applicationServerInfo.IsUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Infof("is service %v application server %v is up, healthcheck to %v is ok",
			serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
			applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
			applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
		applicationServerInfo.IsUp = true
		if err := hc.inclideApplicationServerInIPVS(serviceInfo, applicationServerInfo); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Errorf("Heathcheck error: inclideApplicationServerInIPVS error: %v", err)
		}
		return
	}
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

func (hc *HeathcheckEntity) tcpCheckFail(healthcheckAddress string, timeout time.Duration) bool {
	hcSlice := strings.Split(healthcheckAddress, ":")
	hcIP := hcSlice[0]
	hcPort := hcSlice[1]

	dialer := net.Dialer{
		LocalAddr: hc.techInterface,
		Timeout:   timeout}

	conn, err := dialer.Dial("tcp", net.JoinHostPort(hcIP, hcPort))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Connecting tcp connect error: %v", err)
		return true
	}
	defer conn.Close()

	if conn != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck info port opened: %v", net.JoinHostPort(hcIP, hcPort))
		return false
	}

	// somehow it can be..
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Error("Heathcheck has unknown error: connection is nil, but have no errors")
	return true
}

func (hc *HeathcheckEntity) httpCheckFail(healthcheckAddress string, timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Connecting http error: %v", err)
		return true
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return true
	}
	return false
}

func (hc *HeathcheckEntity) icmpCheckFail(healthcheckAddress string, timeout time.Duration) bool {
	// Start listening for icmp replies
	icpmConnection, err := icmp.ListenPacket("ip4:icmp", hc.techInterface.String())
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm connection error: %v", err)
		return true
	}
	defer icpmConnection.Close()

	// Get the real IP of the target
	dst, err := net.ResolveIPAddr("ip4", healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm resolve ip addr error: %v", err)
		return true
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
		return true
	}

	// Send it
	n, err := icpmConnection.WriteTo(b, dst)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm write bytes to error: %v", err)
		return true
	} else if n != len(b) {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm write bytes to error (not all of bytes was send): %v", err)
		return true
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = icpmConnection.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm set read deadline error: %v", err)
		return true
	}
	n, peer, err := icpmConnection.ReadFrom(reply)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm read reply error: %v", err)
		return true
	}

	// Let's look what we have in reply
	rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm parse message error: %v", err)
		return true
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck icpm for %v succes", healthcheckAddress)
		return false
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: icpm for %v reply type error: got %+v from %v; want echo reply",
			healthcheckAddress,
			rm,
			peer)
		return true
	}
}

func (hc *HeathcheckEntity) excludeApplicationServerFromIPVS(allServiceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) error {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	return hc.ipvsadm.RemoveApplicationServersFromService(formedServiceData, healthcheckUUID)
}

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(allServiceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) error {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	return hc.ipvsadm.AddApplicationServersForService(formedServiceData, healthcheckUUID)

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
func (hc *HeathcheckEntity) httpAdvancedCheckFail(serverHealthcheck domain.ServerHealthcheck,
	timeout time.Duration) bool {
	switch serverHealthcheck.TypeOfCheck {
	case "http-advanced-json":
		return hc.httpAdvancedJSONCheckFail(serverHealthcheck, timeout)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: http advanced check fail error: unknown check type: %v", serverHealthcheck.TypeOfCheck)
		return true
	}
}

// http advanced json start
func (hc *HeathcheckEntity) httpAdvancedJSONCheckFail(serverHealthcheck domain.ServerHealthcheck,
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
		return true
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Connecting http advanced JSON check error: %v", err)
		return true
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return true
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
			return true
		}
	}

	if u.UnknowMap == nil && u.UnknowArrayOfMaps == nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Tracef("Heathcheck error: http advanced JSON check fail error: response is nil from: %v", serverHealthcheck.HealthcheckAddress)
		return true
	}

	for _, aHP := range serverHealthcheck.AdvancedHealthcheckParameters { // go through the array of search objects
		if aHP.NearFieldsMode { // mode for finding all matches for the desired object in a single map
			if hc.isFinderForNearFieldsModeFail(aHP.UserDefinedData, u, serverHealthcheck.HealthcheckAddress) { // if false do not return, continue range params
				return true
			}
		} else {
			if hc.isFinderMapToMapFail(aHP.UserDefinedData, u, serverHealthcheck.HealthcheckAddress) { // if false do not return, continue range params
				return true
			}
		}
	}
	return false
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
