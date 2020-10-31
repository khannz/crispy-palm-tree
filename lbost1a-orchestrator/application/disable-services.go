package application

import (
	"fmt"
	"strings"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/usecase"
	"github.com/sirupsen/logrus"
)

// DisableRuntimeSettings ...
func (balancerFacade *BalancerFacade) DisableRuntimeSettings(isMockMode bool, id string) error {
	balancerFacade.Locker.Lock()
	defer balancerFacade.Locker.Unlock()
	var errors []error
	servicesConfigsFromStorage, err := balancerFacade.CacheStorage.LoadAllStorageDataToDomainModels()
	if err != nil {
		balancerFacade.Logging.WithFields(logrus.Fields{"event id": id}).Warnf("fail to load storage config when programm stop: %v", err)
		return err
	}
	// task: removeTunnels for all services
	for _, serviceConfigFromStorage := range servicesConfigsFromStorage {
		if err := balancerFacade.DisableRemoveService(serviceConfigFromStorage, isMockMode, id); err != nil {
			balancerFacade.Logging.WithFields(logrus.Fields{"event id": id}).Warnf("can't remove service when programm stop: %v", err)
			errors = append(errors, err)
		}
	}
	// to make sure that the ipvs is cleared
	return combineErrors(errors)
}

// DisableRemoveService ...
func (balancerFacade *BalancerFacade) DisableRemoveService(serviceConfigFromStorage *domain.ServiceInfo,
	isMockMode bool,
	id string) error {
	var errors []error
	tunnelsFilesInfo := usecase.FormTunnelsFilesInfo(serviceConfigFromStorage.ApplicationServers, balancerFacade.CacheStorage)

	if err := balancerFacade.TunnelConfig.RemoveAllTunnels(tunnelsFilesInfo, id); err != nil {
		balancerFacade.Logging.WithFields(logrus.Fields{"event id": id}).Warnf("can't remove tunnels: %v", err)
		errors = append(errors, err)
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
