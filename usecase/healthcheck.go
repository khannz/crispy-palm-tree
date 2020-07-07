package usecase

// TODO: healthchecks != usecase!
import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const healthcheckName = "healthcheck"
const healthcheckUUID = "00000000-0000-0000-0000-000000000004"

type failedServices struct {
	sync.Mutex
	wg    *sync.WaitGroup
	count int
}

type dummyWorker struct {
	sync.Mutex
}

// HeathcheckEntity ...
type HeathcheckEntity struct {
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	ipvsadm           *portadapter.IPVSADMEntity // so dirty
	techInterface     *net.TCPAddr
	locker            *domain.Locker
	gracefullShutdown *domain.GracefullShutdown
	signalChan        chan os.Signal
	waitRetry         chan struct{}
	dw                *dummyWorker
	isMockMode        bool
	logging           *logrus.Logger
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	ipvsadm *portadapter.IPVSADMEntity,
	rawTechInterface string,
	locker *domain.Locker,
	gracefullShutdown *domain.GracefullShutdown,
	signalChan chan os.Signal,
	isMockMode bool,
	logging *logrus.Logger) *HeathcheckEntity {
	ti, _, _ := net.ParseCIDR(rawTechInterface + "/32")

	return &HeathcheckEntity{
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		ipvsadm:           ipvsadm,
		techInterface:     &net.TCPAddr{IP: ti},
		locker:            locker,
		gracefullShutdown: gracefullShutdown,
		signalChan:        signalChan,
		waitRetry:         make(chan struct{}),
		dw:                new(dummyWorker),
		isMockMode:        isMockMode,
		logging:           logging,
	}
}

// StartHealthchecks ...
func (hc *HeathcheckEntity) StartHealthchecks(checkTime time.Duration) {
	ticker := time.NewTicker(checkTime)
	for {
		select {
		case <-ticker.C:
			hc.tryHC()
		case <-hc.waitRetry:
			hc.tryHC()
		}
	}
}

func (hc *HeathcheckEntity) tryHC() {
	// gracefull shutdown part start
	hc.locker.Lock()
	defer hc.locker.Unlock()
	hc.gracefullShutdown.Lock()
	if hc.gracefullShutdown.ShutdownNow {
		defer hc.gracefullShutdown.Unlock()
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Warnf("program got shutdown signal, job heathcheck cancel")
		return
	}
	hc.gracefullShutdown.UsecasesJobs++
	hc.gracefullShutdown.Unlock()
	defer decreaseJobs(hc.gracefullShutdown)
	// gracefull shutdown part end

	servicesInfo, err := hc.cacheStorage.LoadAllStorageDataToDomainModel()
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Info("Fail to load storage data to services info for healthcheck")
		go func(hc *HeathcheckEntity) {
			time.Sleep(5 * time.Second)
			hc.waitRetry <- struct{}{}
		}(hc)
		return
	}
	hc.CheckAllApplicationServersInServices(servicesInfo)
}

// CheckAllApplicationServersInServices wg here - because waiting for the end of checks leads to a top-level unlock of execution threads
func (hc *HeathcheckEntity) CheckAllApplicationServersInServices(servicesInfo []*domain.ServiceInfo) {
	serviceWG := new(sync.WaitGroup)
	for _, serviceInfo := range servicesInfo {
		serviceWG.Add(1)
		hc.CheckApplicationServersInService(serviceInfo, serviceWG)
	}
	serviceWG.Wait()
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(serviceInfo *domain.ServiceInfo, serviceWG *sync.WaitGroup) {
	fs := &failedServices{wg: new(sync.WaitGroup)}
	for _, applicationServerInfo := range serviceInfo.ApplicationServers {
		fs.wg.Add(1)
		go hc.checkApplicationServerInService(serviceInfo,
			applicationServerInfo,
			fs)
	}
	fs.wg.Wait()
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Debugf("Heathcheck: in service %v:%v failed services is %v of %v",
		serviceInfo.ServiceIP,
		serviceInfo.ServicePort,
		fs.count,
		len(serviceInfo.ApplicationServers))
	if len(serviceInfo.ApplicationServers)-fs.count < 1 { // FIXME: hardcode
		serviceInfo.State = false
		hc.removeFromDummyWrapper(serviceInfo.ServiceIP)
	} else {
		serviceInfo.State = true
		hc.addToDummyWrapper(serviceInfo.ServiceIP)
	}
	hc.updateInStorages(serviceInfo)
	serviceWG.Done()
}

func (hc *HeathcheckEntity) removeFromDummyWrapper(serviceIP string) {
	hc.dw.Lock()
	defer hc.dw.Unlock()
	if !hc.isMockMode {
		if err := RemoveFromDummy(serviceIP); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Errorf("Heathcheck error: can't remove service ip from dummy: %v", err)
		}
	}
}

func (hc *HeathcheckEntity) addToDummyWrapper(serviceIP string) {
	hc.dw.Lock()
	defer hc.dw.Unlock()
	if !hc.isMockMode {
		if err := addToDummy(serviceIP); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Errorf("Heathcheck error: can't add service ip to dummy: %v", err)
		}
	}
}

func (hc *HeathcheckEntity) checkApplicationServerInService(serviceInfo *domain.ServiceInfo,
	applicationServerInfo *domain.ApplicationServer,
	fs *failedServices) {
	defer fs.wg.Done()
	switch serviceInfo.Healthcheck.Type {
	case "tcp":
		if hc.tcpCheckFail(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout) {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if applicationServerInfo.State {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down, healthcheck to %v fail",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
					applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
				applicationServerInfo.State = false
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: excludeApplicationServerFromIPVS error: %v", err)
				}
			}
			return
		}
	case "http":
		if hc.httpCheckFail(applicationServerInfo.ServerHealthcheck.HealthcheckAddress,
			serviceInfo.Healthcheck.Timeout) {
			fs.Lock()
			fs.count++
			fs.Unlock()
			if applicationServerInfo.State {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down, healthcheck to %v fail",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
					applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
				applicationServerInfo.State = false
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: excludeApplicationServerFromIPVS error: %v", err)
				}
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

	if !applicationServerInfo.State {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Infof("is service %v application server %v is up, healthcheck to %v is ok",
			serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
			applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort,
			applicationServerInfo.ServerHealthcheck.HealthcheckAddress)
		applicationServerInfo.State = true
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
		}).Debugf("Heathcheck error: Connecting tcp connect error: %v", err)
		return true
	}
	defer conn.Close()

	if conn != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Debugf("Heathcheck info port opened: %v", net.JoinHostPort(hcIP, hcPort))
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
		}).Debugf("Heathcheck error: Connecting http error: %v", err)
		return true
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Debugf("Heathcheck error: Read http response errror: %v", err)
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

func (hc *HeathcheckEntity) updateApplicationServicesState(serviceInfo *domain.ServiceInfo) error {
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
						applicationServiceInfo.State = currentApplicationServer.State
					}
				}
			}
			return nil
		}
	}
	return fmt.Errorf("service %v:%v not found in current services", serviceInfo.ServiceIP, serviceInfo.ServicePort)
}

func addToDummy(serviceIP string) error {
	addrs, err := getDummyAddrs()
	if err != nil {
		return err
	}
	var addrIsFounded bool
	incomeIPAndMask := serviceIP + "/32"
	for _, addr := range addrs {
		if incomeIPAndMask == addr.String() {
			addrIsFounded = true
			break

		}
	}
	if !addrIsFounded {
		if err := addAddr(incomeIPAndMask); err != nil {
			return fmt.Errorf("can't add ip addr %v, got err %v", incomeIPAndMask, err)
		}
	}
	return nil
}

// RemoveFromDummy remove service from dummy
func RemoveFromDummy(serviceIP string) error {
	addrs, err := getDummyAddrs()
	if err != nil {
		return err
	}
	var addrIsFounded bool
	incomeIPAndMask := serviceIP + "/32"
	for _, addr := range addrs {
		if incomeIPAndMask == addr.String() {
			addrIsFounded = true
			break
		}
	}
	if addrIsFounded {
		if err := removeAddr(incomeIPAndMask); err != nil {
			return fmt.Errorf("can't remove ip addr %v, got err %v", incomeIPAndMask, err)
		}
	}
	return nil
}

func getDummyAddrs() ([]net.Addr, error) {
	i, err := net.InterfaceByName("dummy0")
	if err != nil {
		return nil, fmt.Errorf("can't get InterfaceByName: %v", err)
	}
	addrs, err := i.Addrs()
	if err != nil {
		return nil, fmt.Errorf("can't addrs in interfaces: %v", err)
	}
	return addrs, err
}

func removeAddr(addrForDel string) error {
	dummy, err := netlink.LinkByName("dummy0") // hardcoded
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(addrForDel)
	if err = netlink.AddrDel(dummy, addr); err != nil {
		return err
	}
	return nil
}

func addAddr(addrForAdd string) error {
	dummy, err := netlink.LinkByName("dummy0") // hardcoded
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(addrForAdd)
	if err = netlink.AddrAdd(dummy, addr); err != nil {
		return err
	}
	return nil
}
