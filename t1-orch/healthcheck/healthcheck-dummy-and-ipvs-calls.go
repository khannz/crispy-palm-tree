package healthcheck

import (
	"fmt"

	domain "github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

func (hc *HealthcheckEntity) announceLogic(serviceIP string, newServiceStateIsUp bool, id string) {
	hc.Lock() // TODO need lock only dummy map, not all HealthcheckEntity
	defer hc.Unlock()
	// paranoid checks key start
	if _, inMap := hc.dummyEntities[serviceIP]; !inMap {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Errorf("critical error: somehow key %v not in map %v", serviceIP, hc.dummyEntities)
		return
	}
	// paranoid checks key end

	if newServiceStateIsUp {
		hc.dummyEntities[serviceIP].announcedForDummy++
		if hc.dummyEntities[serviceIP].totalForDummy == hc.dummyEntities[serviceIP].announcedForDummy {
			if !hc.dummyEntities[serviceIP].isAnnouncedAtDummy { // not announced yet
				hc.dummyEntities[serviceIP].isAnnouncedAtDummy = true
				if err := hc.dw.AddToDummy(serviceIP, id); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": id,
					}).Errorf("fail announce service %v: %v", serviceIP, err)
				}
				hc.logging.WithFields(logrus.Fields{
					"entity":   healthcheckName,
					"event id": id,
				}).Infof("service %v announced", serviceIP)
			}
		}
	} else {
		hc.dummyEntities[serviceIP].announcedForDummy--
		if hc.dummyEntities[serviceIP].isAnnouncedAtDummy { // announced now
			hc.dummyEntities[serviceIP].isAnnouncedAtDummy = false
			if err := hc.dw.RemoveFromDummy(serviceIP, id); err != nil {
				hc.logging.WithFields(logrus.Fields{
					"entity":   healthcheckName,
					"event id": id,
				}).Errorf("fail remove announce for service %v: %v", serviceIP, err)
			}
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": id,
			}).Infof("service %v stop announced", serviceIP)
		}
	}
}

func (hc *HealthcheckEntity) addServiceToIPVS(hcService *domain.ServiceInfo, id string) error {
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

func (hc *HealthcheckEntity) removeServiceFromIPVS(hcService *domain.ServiceInfo, id string) error {
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

func (hc *HealthcheckEntity) inclideApplicationServerInIPVS(hcService *domain.ServiceInfo,
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

func (hc *HealthcheckEntity) excludeApplicationServerFromIPVS(hcService *domain.ServiceInfo,
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
