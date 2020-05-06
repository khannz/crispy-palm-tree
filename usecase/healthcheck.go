package usecase

// TODO: healthchecks != usecase!
import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"
const healthcheckUUID = "00000000-0000-0000-0000-000000000004"

// HeathcheckEntity ...
type HeathcheckEntity struct {
	cacheStorage      *portadapter.StorageEntity // so dirty
	persistentStorage *portadapter.StorageEntity // so dirty
	configuratorVRRP  domain.ServiceWorker
	techInterface     *net.TCPAddr
	locker            *domain.Locker
	gracefullShutdown *domain.GracefullShutdown
	signalChan        chan os.Signal
	waitRetry         chan struct{}
	logging           *logrus.Logger
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(cacheStorage *portadapter.StorageEntity,
	persistentStorage *portadapter.StorageEntity,
	configuratorVRRP domain.ServiceWorker,
	rawTechInterface string,
	locker *domain.Locker,
	gracefullShutdown *domain.GracefullShutdown,
	signalChan chan os.Signal,
	logging *logrus.Logger) *HeathcheckEntity {
	ti, _, _ := net.ParseCIDR(rawTechInterface + "/32")

	return &HeathcheckEntity{
		cacheStorage:      cacheStorage,
		persistentStorage: persistentStorage,
		configuratorVRRP:  configuratorVRRP,
		techInterface:     &net.TCPAddr{IP: ti},
		locker:            locker,
		gracefullShutdown: gracefullShutdown,
		signalChan:        signalChan,
		waitRetry:         make(chan struct{}),
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

// CheckAllApplicationServersInServices ...
func (hc *HeathcheckEntity) CheckAllApplicationServersInServices(servicesInfo []*domain.ServiceInfo) {
	for _, serviceInfo := range servicesInfo {
		hc.CheckApplicationServersInService(serviceInfo)
	}
}

// CheckApplicationServersInService ...
func (hc *HeathcheckEntity) CheckApplicationServersInService(serviceInfo *domain.ServiceInfo) {
	if err := hc.updateApplicationServicesState(serviceInfo); err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: update application services state error: %v", err)
		return
	}
	var failedServices int
	for _, applicationServerInfo := range serviceInfo.ApplicationServers {
		if hc.tcpCheckFail(applicationServerInfo.ServerIP, applicationServerInfo.ServerPort) {
			failedServices++
			if applicationServerInfo.State {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Infof("is service %v application server %v is down",
					serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
					applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort)
				applicationServerInfo.State = false
				if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":     healthcheckName,
						"event uuid": healthcheckUUID,
					}).Errorf("Heathcheck error: RemoveApplicationServersFromService error: %v", err)
				}
			}
			continue
		}
		if !applicationServerInfo.State {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Infof("is service %v application server %v is up",
				serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
				applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort)
			applicationServerInfo.State = true
			if err := hc.inclideApplicationServerInIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: AddApplicationServersForService error: %v", err)
			}
			continue
		}
	}
	if len(serviceInfo.ApplicationServers)-failedServices < 2 { // hardcode
		serviceInfo.State = false
	} else if !serviceInfo.State {
		serviceInfo.State = true
	}
	hc.updateInStorages(serviceInfo)
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

func (hc *HeathcheckEntity) tcpCheckFail(ip, port string) bool {
	timeout := time.Second * 2
	// FIXME: refactor that
	nip, _ := strconv.Atoi(port)
	nip += 10000
	np := strconv.Itoa(nip)

	dialer := net.Dialer{
		LocalAddr: hc.techInterface,
		Timeout:   timeout}

	conn, err := dialer.Dial("tcp", net.JoinHostPort(ip, np))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Debugf("Heathcheck error: Connecting error: %v", err)
		return true
	}
	defer conn.Close()

	if conn != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Debugf("Heathcheck info port opened: %v", net.JoinHostPort(ip, np))
		return false
	}

	// somehow it can be..
	hc.logging.WithFields(logrus.Fields{
		"entity":     healthcheckName,
		"event uuid": healthcheckUUID,
	}).Error("Heathcheck has unknown error: connection is nil, but have no errors")
	return true
}

func (hc *HeathcheckEntity) excludeApplicationServerFromIPVS(allServiceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) error {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	return hc.configuratorVRRP.RemoveApplicationServersFromService(formedServiceData, healthcheckUUID)
}

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(allServiceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) error {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	return hc.configuratorVRRP.AddApplicationServersForService(formedServiceData, healthcheckUUID)

}

// AtStartCheckAllApplicationServersInServices ...
func (hc *HeathcheckEntity) AtStartCheckAllApplicationServersInServices(servicesInfo []*domain.ServiceInfo) {
	for _, serviceInfo := range servicesInfo {
		hc.AtStartCheckApplicationServersInService(serviceInfo)
	}
}

// AtStartCheckApplicationServersInService ...
func (hc *HeathcheckEntity) AtStartCheckApplicationServersInService(serviceInfo *domain.ServiceInfo) {
	var failedServices int
	for _, applicationServerInfo := range serviceInfo.ApplicationServers {
		if hc.tcpCheckFail(applicationServerInfo.ServerIP, applicationServerInfo.ServerPort) {
			failedServices++
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Infof("is service %v application server %v is down",
				serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
				applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort)
			applicationServerInfo.State = false
			if err := hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: RemoveApplicationServersFromService error: %v", err)
			}
			continue
		}
		if !applicationServerInfo.State {
			hc.logging.WithFields(logrus.Fields{
				"entity":     healthcheckName,
				"event uuid": healthcheckUUID,
			}).Infof("is service %v application server %v is up",
				serviceInfo.ServiceIP+":"+serviceInfo.ServicePort,
				applicationServerInfo.ServerIP+":"+applicationServerInfo.ServerPort)
			applicationServerInfo.State = true
			if err := hc.inclideApplicationServerInIPVS(serviceInfo, applicationServerInfo); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":     healthcheckName,
					"event uuid": healthcheckUUID,
				}).Errorf("Heathcheck error: AddApplicationServersForService error: %v", err)
			}
			continue
		}
	}
	if len(serviceInfo.ApplicationServers)-failedServices < 2 { // hardcode
		serviceInfo.State = false
	} else if !serviceInfo.State {
		serviceInfo.State = true
	}
	hc.updateInStorages(serviceInfo)
}

func (hc *HeathcheckEntity) updateApplicationServicesState(serviceInfo *domain.ServiceInfo) error {
	currentConfig, err := hc.configuratorVRRP.ReadCurrentConfig()
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
