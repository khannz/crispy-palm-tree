package healthcheck

import (
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/sirupsen/logrus"
)

func (hc *HeathcheckEntity) startHealthchecksForCurrentService(hcService *domain.ServiceInfo) {
	// first run hc at create entity
	hc.CheckApplicationServersInService(hcService) // lock hc, hcService, dummy
	hc.logging.Tracef("hc service: %v", hcService)
	ticker := time.NewTicker(hcService.HelloTimer)
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

// CheckApplicationServersInService ... TODO: !!! rename that. not only checks here, also set service state
func (hc *HeathcheckEntity) CheckApplicationServersInService(hcService *domain.ServiceInfo) {
	defer hcService.FailedApplicationServers.SetFailedApplicationServersToZero()
	newID := hc.idGenerator.NewID()
	for k := range hcService.ApplicationServers {
		hcService.FailedApplicationServers.Wg.Add(1)
		go hc.checkApplicationServerInService(hcService,
			k,
			newID) // lock hcService
	}
	hcService.FailedApplicationServers.Wg.Wait()
	percentageUp := percentageOfUp(len(hcService.ApplicationServers), hcService.FailedApplicationServers.Count)
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": newID,
	}).Tracef("Heathcheck: in service %v failed services is %v of %v; %v up percent of %v max for this service",
		hcService.Address,
		hcService.FailedApplicationServers.Count,
		len(hcService.ApplicationServers),
		percentageUp,
		hcService.Quorum)
	isServiceUp := percentageOfDownBelowMPercentOfAlivedForUp(percentageUp, hcService.Quorum)
	hc.logging.Tracef("Old service state %v. New service state %v", hcService.IsUp, isServiceUp)

	if !hcService.IsUp && isServiceUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": newID,
		}).Warnf("service %v is up now", hcService.Address)
		hcService.IsUp = true
		hc.annonceLogic(hcService.IP, hcService.IsUp, newID) // lock hc and dummy
	} else if hcService.IsUp && !isServiceUp {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": newID,
		}).Warnf("service %v is down now", hcService.Address)
		hcService.IsUp = false
		hc.annonceLogic(hcService.IP, hcService.IsUp, newID) // lock hc and dummy
	} else {
		hc.logging.Tracef("service state not changed: is up: %v", hcService.IsUp)
	}
	// hc.updateInStorage(hcService, newID)
}

func (hc *HeathcheckEntity) checkApplicationServerInService(hcService *domain.ServiceInfo,
	applicationServerInfoKey string,
	id string) {
	// TODO: still can be refactored
	defer hcService.FailedApplicationServers.Wg.Done()
	isCheckOk := hc.isApplicationServerOkNow(hcService, applicationServerInfoKey, id)

	hc.moveApplicationServerStateIndexes(hcService, applicationServerInfoKey, isCheckOk)                                                     // lock hcService
	isApplicationServerUp, isApplicationServerChangeState := hc.isApplicationServerUpAndStateChange(hcService, applicationServerInfoKey, id) // lock hcService
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": id,
	}).Tracef("for server %v: isCheckOk: %v, isApplicationServerUp: %v, isApplicationServerChangeState: %v ",
		hcService.ApplicationServers[applicationServerInfoKey].Address,
		isCheckOk,
		isApplicationServerUp,
		isApplicationServerChangeState)
	if !isCheckOk {
		hc.logging.Tracef("one hc for application server %v is up: %v; is server change state: %v",
			hcService.ApplicationServers[applicationServerInfoKey].Address,
			isApplicationServerUp,
			isApplicationServerChangeState)
		if !isApplicationServerUp {
			hcService.FailedApplicationServers.Lock()
			hcService.FailedApplicationServers.Count++
			hcService.FailedApplicationServers.Unlock()
			if isApplicationServerChangeState {
				if err := hc.excludeApplicationServerFromIPVS(hcService, hcService.ApplicationServers[applicationServerInfoKey], id); err != nil {
					hc.logging.WithFields(logrus.Fields{
						"entity":   healthcheckName,
						"event id": id,
					}).Errorf("Heathcheck error: exclude application server from IPVS: %v", err)
				}
			}
		}
		return
	}

	hc.logging.Tracef("one hc for application server %v is up: %v; is server change state: %v",
		hcService.ApplicationServers[applicationServerInfoKey].Address,
		isApplicationServerUp,
		isApplicationServerChangeState)
	if !isApplicationServerUp {
		hcService.FailedApplicationServers.Lock()
		hcService.FailedApplicationServers.Count++
		hcService.FailedApplicationServers.Unlock()
	}

	if isApplicationServerUp && isApplicationServerChangeState {
		if err := hc.inclideApplicationServerInIPVS(hcService, hcService.ApplicationServers[applicationServerInfoKey], id); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": id,
			}).Errorf("Heathcheck error: inclide application server in IPVS error: %v", err)
		}
		return
	}
}

// TODO: recaftor work for check address, at this moment it looks like magic
func (hc *HeathcheckEntity) isApplicationServerOkNow(hcService *domain.ServiceInfo,
	applicationServerInfoKey string,
	id string) bool {
	switch hcService.HealthcheckType {
	case "tcp":
		return hc.healthcheckChecker.IsTcpCheckOk(hcService.ApplicationServers[applicationServerInfoKey].HealthcheckAddress,
			hcService.ResponseTimer,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)

	case "http":
		return hc.healthcheckChecker.IsHttpCheckOk(hcService.ApplicationServers[applicationServerInfoKey].HealthcheckAddress,
			hcService.ResponseTimer,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	case "https":
		return hc.healthcheckChecker.IsHttpsCheckOk(hcService.ApplicationServers[applicationServerInfoKey].HealthcheckAddress,
			hcService.ResponseTimer,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	case "http-advanced":
		return hc.healthcheckChecker.IsHttpAdvancedCheckOk(hcService.HealthcheckType,
			hcService.ApplicationServers[applicationServerInfoKey].HealthcheckAddress,
			hcService.HCNearFieldsMode,
			hcService.HCUserDefinedData,
			hcService.ResponseTimer,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	case "icmp":
		return hc.healthcheckChecker.IsIcmpCheckOk(hcService.IP,
			hcService.ResponseTimer,
			hcService.ApplicationServers[applicationServerInfoKey].InternalHC.Mark,
			id)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Errorf("Heathcheck error: unknown healtcheck type: %v", hcService.HealthcheckType)
		return false // must never will bfe. all data already validated
	}
}

func (hc *HeathcheckEntity) isApplicationServerUpAndStateChange(hcService *domain.ServiceInfo,
	applicationServerInfoKey string,
	id string) (bool, bool) {
	//return isUp and isChagedState booleans
	hcService.Lock()
	defer hcService.Unlock()
	hc.logging.Tracef("real: %v, RetriesCounterForDown: %v", hcService.ApplicationServers[applicationServerInfoKey].Address, hcService.ApplicationServers[applicationServerInfoKey].InternalHC.AliveThreshold)
	hc.logging.Tracef("real: %v, RetriesCounterForUp: %v", hcService.ApplicationServers[applicationServerInfoKey].Address, hcService.ApplicationServers[applicationServerInfoKey].InternalHC.DeadThreshold)

	if hcService.ApplicationServers[applicationServerInfoKey].IsUp { // !!!
		// check it not down
		for _, isUp := range hcService.ApplicationServers[applicationServerInfoKey].InternalHC.DeadThreshold {
			if isUp {
				return true, false // do not change up state
			}
		}
		hcService.ApplicationServers[applicationServerInfoKey].IsUp = false // if all hc fail at RetriesCounterForDown - change state
		hc.logging.WithFields(logrus.Fields{
			"event id": id,
		}).Tracef("at service %v real server %v DOWN", hcService.Address,
			hcService.ApplicationServers[applicationServerInfoKey].Address)
		return false, true
	}

	for _, isUp := range hcService.ApplicationServers[applicationServerInfoKey].InternalHC.AliveThreshold {
		if !isUp {
			// do not change down state
			return false, false
		}
	}

	// all RetriesCounterForUp true
	hcService.ApplicationServers[applicationServerInfoKey].IsUp = true // if all hc fail at RetriesCounterForDown - change state
	hc.logging.WithFields(logrus.Fields{
		"event id": id,
	}).Tracef("at service %v real server %v UP", hcService.Address,
		hcService.ApplicationServers[applicationServerInfoKey].Address)
	return true, true

}
