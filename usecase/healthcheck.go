package usecase

// TODO: healthchecks != usecase!
import (
	"net"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const healthcheckName = "healthcheck"
const healthcheckUUID = "00000000-0000-0000-0000-000000000004"

// HeathcheckEntity ...
type HeathcheckEntity struct {
	locker           *domain.Locker
	configuratorVRRP domain.ServiceWorker
	logging          *logrus.Logger
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity(locker *domain.Locker,
	configuratorVRRP domain.ServiceWorker,
	logging *logrus.Logger) *HeathcheckEntity {
	return &HeathcheckEntity{
		locker:           locker,
		configuratorVRRP: configuratorVRRP,
		logging:          logging,
	}
}

// CheckApplicationServersInServices ...
func (hc *HeathcheckEntity) CheckApplicationServersInServices(servicesInfo []*domain.ServiceInfo) {
	hc.locker.Lock()
	defer hc.locker.Unlock()
	for _, serviceInfo := range servicesInfo {
		hc.CheckApplicationServersInService(serviceInfo)
	}
}

// CheckApplicationServersInService ...
// TODO: maybe check application server and return err, wheb usecase create service/add app servers?
func (hc *HeathcheckEntity) CheckApplicationServersInService(serviceInfo *domain.ServiceInfo) {
	var failedServices int
	for _, applicationServerInfo := range serviceInfo.ApplicationServers {
		if hc.tcpCheckFail(applicationServerInfo.ServerIP, applicationServerInfo.ServerPort) {
			failedServices++
			if applicationServerInfo.State {
				hc.excludeApplicationServerFromIPVS(serviceInfo, applicationServerInfo)
			}
			applicationServerInfo.State = false
			continue
		}
		if !applicationServerInfo.State {
			hc.inclideApplicationServerInIPVS(serviceInfo, applicationServerInfo)
		}
		applicationServerInfo.State = true
	}
	if len(serviceInfo.ApplicationServers)-failedServices < 2 { // hardcode
		serviceInfo.State = false
	}
	serviceInfo.State = true
}

func (hc *HeathcheckEntity) tcpCheckFail(ip, port string) bool {
	timeout := time.Second * 2
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: Connecting error: %v", err)
		return true
	}
	if conn != nil {
		defer conn.Close()
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Debugf("Heathcheck info port opened: %v", net.JoinHostPort(ip, port))
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
	applicationServer *domain.ApplicationServer) {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	err := hc.configuratorVRRP.RemoveApplicationServersFromService(formedServiceData, healthcheckUUID)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: RemoveApplicationServersFromService error: %v", err)
	}
}

func (hc *HeathcheckEntity) inclideApplicationServerInIPVS(allServiceInfo *domain.ServiceInfo,
	applicationServer *domain.ApplicationServer) {
	formedServiceData := &domain.ServiceInfo{
		ServiceIP:          allServiceInfo.ServiceIP,
		ServicePort:        allServiceInfo.ServicePort,
		ApplicationServers: []*domain.ApplicationServer{applicationServer},
	}
	err := hc.configuratorVRRP.AddApplicationServersForService(formedServiceData, healthcheckUUID)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":     healthcheckName,
			"event uuid": healthcheckUUID,
		}).Errorf("Heathcheck error: AddApplicationServersForService error: %v", err)
	}
}
