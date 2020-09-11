package application

import (
	"fmt"
	"strings"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/usecase"
	"github.com/sirupsen/logrus"
)

// DisableRuntimeSettings ...
func (balancerFacade *BalancerFacade) DisableRuntimeSettings(isMockMode bool, uuid string) error {
	balancerFacade.Locker.Lock()
	defer balancerFacade.Locker.Unlock()
	var errors []error
	servicesConfigsFromStorage, err := balancerFacade.CacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": uuid}).Errorf("fail to load  storage config when programm stop: %v", err)
		return err
	}
	for _, serviceConfigFromStorage := range servicesConfigsFromStorage {
		if err := balancerFacade.DisableRemoveService(serviceConfigFromStorage, isMockMode, uuid); err != nil {
			balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": uuid}).Errorf("can't remove service when programm stop: %v", err)
			errors = append(errors, err)
		}
	}
	// to make sure that the ipvs is cleared
	if err := balancerFacade.IPVSADMConfigurator.Flush(); err != nil {
		balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": uuid}).Errorf("IPVSADM can't flush data when programm stop: %v", err)
		errors = append(errors, err)
	}

	return combineErrors(errors)
}

// DisableRemoveService ...
func (balancerFacade *BalancerFacade) DisableRemoveService(serviceConfigFromStorage *domain.ServiceInfo,
	isMockMode bool,
	uuid string) error {
	var errors []error
	tunnelsFilesInfo := usecase.FormTunnelsFilesInfo(serviceConfigFromStorage.ApplicationServers, balancerFacade.CacheStorage)
	_, err := balancerFacade.TunnelConfig.RemoveTunnels(tunnelsFilesInfo, uuid)
	if err != nil {
		balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": uuid}).Errorf("can't remove tunnels: %v", err)
		errors = append(errors, err)
	}

	vip, port, _, _, protocol, _, err := domain.PrepareDataForIPVS(serviceConfigFromStorage.ServiceIP,
		serviceConfigFromStorage.ServicePort,
		serviceConfigFromStorage.RoutingType,
		serviceConfigFromStorage.BalanceType,
		serviceConfigFromStorage.Protocol,
		serviceConfigFromStorage.ApplicationServers)
	if err != nil {
		return fmt.Errorf("Error prepare data for IPVS: %v", err)
	}
	if err := balancerFacade.IPVSADMConfigurator.RemoveService(vip, port, protocol, uuid); err != nil {
		balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": uuid}).Errorf("ipvsadm can't remove service: %v. got error: %v", serviceConfigFromStorage, err)
		errors = append(errors, err)
	}

	balancerFacade.HeathcheckEntity.RemoveServiceFromHealtchecks(serviceConfigFromStorage)

	if !isMockMode {
		if err := usecase.RemoveFromDummy(serviceConfigFromStorage.ServiceIP); err != nil {
			balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": uuid}).Errorf("can't remove from dummy: %v", err)
			errors = append(errors, err)
		}
	}
	return combineErrors(errors)
}

func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	var errorsStringSlice []string
	for _, err := range errors {
		errorsStringSlice = append(errorsStringSlice, err.Error())
	}
	return fmt.Errorf(strings.Join(errorsStringSlice, "\n"))
}
