package healthcheck

import (
	"fmt"

	domain "github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

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
