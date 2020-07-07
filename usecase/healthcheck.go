package usecase

// TODO: healthchecks != usecase!
import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"
const healthcheckUUID = "00000000-0000-0000-0000-000000000004"

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
	cacheStorage       *portadapter.StorageEntity // so dirty
	persistentStorage  *portadapter.StorageEntity // so dirty
	ipvsadm            *portadapter.IPVSADMEntity // so dirty
	techInterface      *net.TCPAddr
	locker             *domain.Locker
	gracefullShutdown  *domain.GracefullShutdown
	dw                 *dummyWorker
	isMockMode         bool
	logging            *logrus.Logger
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	ipvsadm *portadapter.IPVSADMEntity,
	rawTechInterface string,
	locker *domain.Locker,
	gracefullShutdown *domain.GracefullShutdown,
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
		gracefullShutdown:  gracefullShutdown,
		dw:                 new(dummyWorker),
		isMockMode:         isMockMode,
		logging:            logging,
	}
}

// StartGracefullShutdownControlForHealthchecks ...
func (hc *HeathcheckEntity) StartGracefullShutdownControlForHealthchecks() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			hc.Lock()
			if hc.gracefullShutdown.ShutdownNow {
				for _, serviceInfo := range hc.runningHeathchecks {
					serviceInfo.Lock()
					serviceInfo.Healthcheck.StopChecks <- struct{}{}
					serviceInfo.Unlock()
				}
			}
			hc.Unlock()
		}
	}
}

// StartHealthchecksForCurrentServices ...
func (hc *HeathcheckEntity) StartHealthchecksForCurrentServices() error {
	hc.Lock()
	defer hc.Unlock()
	// if shutdown command at start
	if hc.gracefullShutdown.ShutdownNow {
		return nil
	}
	servicesInfo, err := hc.cacheStorage.LoadAllStorageDataToDomainModel()
	if err != nil {
		return fmt.Errorf("fail when try LoadAllStorageDataToDomainModel: %v", err)
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
	if hc.gracefullShutdown.ShutdownNow {
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
	if hc.gracefullShutdown.ShutdownNow {
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
		}).Error("Heathcheck error: RemoveServiceFromHealtchecks error: service %v:%v not found",
			serviceInfo.ServiceIP,
			serviceInfo.ServicePort)
	}
}

// UpdateServiceAtHealtchecks ...
func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(serviceInfo *domain.ServiceInfo) {
	hc.Lock()
	defer hc.Unlock()
	if hc.gracefullShutdown.ShutdownNow {
		return
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
	} else {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Error("Heathcheck error: UpdateServiceAtHealtchecks error: service %v:%v not found",
			serviceInfo.ServiceIP,
			serviceInfo.ServicePort)
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
	hc.gracefullShutdown.Lock()
	hc.gracefullShutdown.UsecasesJobs++
	hc.gracefullShutdown.Unlock()
	defer decreaseJobs(hc.gracefullShutdown)

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
			hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo)
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
			hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo)
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

func (hc *HeathcheckEntity) updateApplicationServicesIsUp(serviceInfo *domain.ServiceInfo) error {
	currentConfig, err := hc.ipvsadm.ReadCurrentConfig()
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("can't read current config: %v", err)
		return nil
	}
	for _, currentServiceInfo := range currentConfig {
		if currentServiceInfo.ServiceIP == serviceInfo.ServiceIP &&
			currentServiceInfo.ServicePort == serviceInfo.ServicePort {
			for _, applicationServiceInfo := range serviceInfo.ApplicationServers {
				for _, currentApplicationServer := range currentServiceInfo.ApplicationServers {
					if applicationServiceInfo.ServerIP == currentApplicationServer.ServerIP &&
						applicationServiceInfo.ServerPort == currentApplicationServer.ServerPort {
						applicationServiceInfo.IsUp = currentApplicationServer.IsUp
					}
				}
			}
			return nil
		}
	}
	return fmt.Errorf("service %v:%v not found in current services", serviceInfo.ServiceIP, serviceInfo.ServicePort)
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
