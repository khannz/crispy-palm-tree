package application

import (
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/khannz/crispy-palm-tree/t1-orch/usecase"
	"github.com/sirupsen/logrus"
)

// T1OrchFacade struct
type T1OrchFacade struct {
	MemoryWorker     domain.MemoryWorker
	RouteWorker      domain.RouteWorker
	HeathcheckEntity *healthcheck.HeathcheckEntity
	GracefulShutdown *domain.GracefulShutdown
	IDgenerator      domain.IDgenerator
	RuntimeServices  map[string]*domain.ServiceInfo
	Logging          *logrus.Logger
}

// NewT1OrchFacade ...
func NewT1OrchFacade(memoryWorker domain.MemoryWorker,
	routeWorker domain.RouteWorker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *T1OrchFacade {

	return &T1OrchFacade{
		MemoryWorker:     memoryWorker,
		RouteWorker:      routeWorker,
		HeathcheckEntity: hc,
		GracefulShutdown: gracefulShutdown,
		IDgenerator:      idGenerator,
		RuntimeServices:  make(map[string]*domain.ServiceInfo),
		Logging:          logging,
	}
}

func (t1OrchFacade *T1OrchFacade) ApplyNewConfig(updatedServicesInfo map[string]*domain.ServiceInfo) error {
	id := t1OrchFacade.IDgenerator.NewID()
	// form diff for runtime config
	servicesForCreate, servicesForUpdate, servicesForRemove := t1OrchFacade.formDiffForNewConfig(updatedServicesInfo)

	// TODO: usecases in gorutines
	if len(servicesForCreate) != 0 {
		if err := t1OrchFacade.CreateServices(servicesForCreate, id); err != nil {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":   consulWorkerName,
				"event id": id,
			}).Errorf("create services error: %v", err)
		}
	}

	if len(servicesForUpdate) != 0 {
		// TODO: implement
	}

	if len(servicesForRemove) != 0 {
		if err := t1OrchFacade.RemoveServices(servicesForRemove, id); err != nil {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":   consulWorkerName,
				"event id": id,
			}).Errorf("remove services error: %v", err)
		}
	}

	return nil
}

func (t1OrchFacade *T1OrchFacade) CreateServices(servicesForCreate map[string]*domain.ServiceInfo,
	id string) error {
	newNewServiceEntity := usecase.NewNewServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, serviceForCreate := range servicesForCreate {
		if err := t1OrchFacade.MemoryWorker.AddService(serviceForCreate); err != nil {
			return err
		}
		if err := newNewServiceEntity.NewService(serviceForCreate, id); err != nil {
			return err
		}
	}
	return nil
}

func (t1OrchFacade *T1OrchFacade) RemoveServices(servicesForRemove map[string]*domain.ServiceInfo,
	id string) error {
	newRemoveServiceEntity := usecase.NewRemoveServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, serviceForCreate := range servicesForRemove {
		if err := t1OrchFacade.MemoryWorker.RemoveService(serviceForCreate); err != nil {
			return err
		}
		if err := newRemoveServiceEntity.RemoveService(serviceForCreate, id); err != nil {
			return err
		}
	}
	return nil
}

func (t1OrchFacade *T1OrchFacade) formDiffForNewConfig(updatedServicesInfo map[string]*domain.ServiceInfo) (map[string]*domain.ServiceInfo, map[string]*domain.ServiceInfo, map[string]*domain.ServiceInfo) {
	servicesForCreate := make(map[string]*domain.ServiceInfo)
	servicesForUpdate := make(map[string]*domain.ServiceInfo)
	servicesForRemove := make(map[string]*domain.ServiceInfo)

	currentServices := t1OrchFacade.MemoryWorker.GetServices()

	for updatedServiceInfoAddress, updatedServiceInfo := range updatedServicesInfo {
		if _, isServiceIn := currentServices[updatedServiceInfoAddress]; isServiceIn {
			servicesForUpdate[updatedServiceInfoAddress] = updatedServiceInfo
		} else {
			servicesForCreate[updatedServiceInfoAddress] = updatedServiceInfo
		}
	}

	for memServiceAddress, memServiceInfo := range currentServices {
		if _, isServiceIn := updatedServicesInfo[memServiceAddress]; !isServiceIn {
			servicesForRemove[memServiceAddress] = memServiceInfo
		}
	}

	return servicesForCreate, servicesForUpdate, servicesForRemove
}
