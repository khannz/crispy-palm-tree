package application

import (
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/khannz/crispy-palm-tree/t1-orch/usecase"
	"github.com/sirupsen/logrus"
)

// TODO: agregate errors in facade

const facadeApplyName = "facade apply config"

// T1OrchFacade struct
type T1OrchFacade struct {
	MemoryWorker     domain.MemoryWorker
	TunnelWorker     domain.TunnelWorker
	RouteWorker      domain.RouteWorker
	HeathcheckEntity *healthcheck.HeathcheckEntity
	GracefulShutdown *domain.GracefulShutdown
	IDgenerator      domain.IDgenerator
	RuntimeServices  map[string]*domain.ServiceInfo
	Logging          *logrus.Logger
}

// NewT1OrchFacade ...
func NewT1OrchFacade(memoryWorker domain.MemoryWorker,
	tunnelWorker domain.TunnelWorker,
	routeWorker domain.RouteWorker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *T1OrchFacade {

	return &T1OrchFacade{
		MemoryWorker:     memoryWorker,
		TunnelWorker:     tunnelWorker,
		RouteWorker:      routeWorker,
		HeathcheckEntity: hc,
		GracefulShutdown: gracefulShutdown,
		IDgenerator:      idGenerator,
		RuntimeServices:  make(map[string]*domain.ServiceInfo),
		Logging:          logging,
	}
}

func (t1OrchFacade *T1OrchFacade) ApplyNewConfig(updatedServicesInfo map[string]*domain.ServiceInfo) error {
	start := time.Now()
	id := t1OrchFacade.IDgenerator.NewID()
	// form diff for runtime config
	servicesForCreate, servicesForUpdate, servicesForRemove := t1OrchFacade.formDiffForNewConfig(updatedServicesInfo)
	t1OrchFacade.Logging.WithFields(logrus.Fields{
		"entity":   facadeApplyName,
		"event id": id,
	}).Infof("got apply new config job: create: %v; update %v; remove: %v",
		len(servicesForCreate),
		len(servicesForUpdate),
		len(servicesForRemove))

	t1OrchFacade.Logging.WithFields(logrus.Fields{
		"entity":   facadeApplyName,
		"event id": id,
	}).Tracef("services for create(%v): %v\nservices for update(%v): %v\nservices for remove(%v): %v",
		servicesForCreate,
		len(servicesForCreate),
		servicesForUpdate,
		len(servicesForUpdate),
		servicesForRemove,
		len(servicesForRemove))

	// TODO: usecases in gorutines
	if len(servicesForCreate) != 0 {
		if err := t1OrchFacade.CreateServices(servicesForCreate, id); err != nil {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":   facadeApplyName,
				"event id": id,
			}).Errorf("create services error: %v", err)
		} else {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":                       consulWorkerName,
				"duration for create services": time.Since(start),
			}).Infof("create %v services done", len(servicesForCreate))
		}
	}

	if len(servicesForUpdate) != 0 {
		if err := t1OrchFacade.UpdateServices(servicesForUpdate, id); err != nil {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":   facadeApplyName,
				"event id": id,
			}).Errorf("update services error: %v", err)
		} else {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":                       consulWorkerName,
				"duration for create services": time.Since(start),
			}).Infof("update %v services done", len(servicesForUpdate))
		}
	}

	if len(servicesForRemove) != 0 {
		if err := t1OrchFacade.RemoveServices(servicesForRemove, id); err != nil {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":   facadeApplyName,
				"event id": id,
			}).Errorf("remove services error: %v", err)
		} else {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":                       consulWorkerName,
				"duration for create services": time.Since(start),
			}).Infof("remove %v services done", len(servicesForRemove))
		}
	}
	t1OrchFacade.Logging.WithFields(logrus.Fields{
		"entity":                        consulWorkerName,
		"duration for apply new config": time.Since(start),
	}).Info("apply config done")
	return nil
}

func (t1OrchFacade *T1OrchFacade) CreateServices(servicesForCreate map[string]*domain.ServiceInfo,
	id string) error {
	newNewServiceEntity := usecase.NewNewServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, serviceForCreate := range servicesForCreate {
		if err := newNewServiceEntity.NewService(serviceForCreate, id); err != nil {
			return err
		}
	}
	return nil
}

func (t1OrchFacade *T1OrchFacade) RemoveServices(servicesForRemove map[string]*domain.ServiceInfo,
	id string) error {
	newRemoveServiceEntity := usecase.NewRemoveServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, serviceForRemove := range servicesForRemove {
		if err := newRemoveServiceEntity.RemoveService(serviceForRemove, id); err != nil {
			return err
		}
	}
	return nil
}

func (t1OrchFacade *T1OrchFacade) UpdateServices(servicesForUpdate map[string]*domain.ServiceInfo,
	id string) error {
	newUpdateServiceEntity := usecase.NewUpdateServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, serviceForUpdate := range servicesForUpdate {
		if err := newUpdateServiceEntity.UpdateService(serviceForUpdate, id); err != nil {
			return err
		}
	}
	return nil
}

func (t1OrchFacade *T1OrchFacade) RemoveAllConfig() error {
	start := time.Now()
	id := t1OrchFacade.IDgenerator.NewID()
	servicesForRemove := t1OrchFacade.MemoryWorker.GetServices()
	t1OrchFacade.Logging.WithFields(logrus.Fields{
		"entity":   facadeApplyName,
		"event id": id,
	}).Infof("got remove all config job: configs for remove %v",
		len(servicesForRemove))
	if servicesForRemove == nil {
		return nil
	}
	newRemoveServiceEntity := usecase.NewRemoveServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, serviceForRemove := range servicesForRemove {
		if err := newRemoveServiceEntity.RemoveService(serviceForRemove, id); err != nil {
			t1OrchFacade.Logging.WithFields(logrus.Fields{
				"entity":   facadeApplyName,
				"event id": id,
			}).Errorf("remove service error: %v", err)
		}
	}
	t1OrchFacade.Logging.WithFields(logrus.Fields{
		"entity":                        consulWorkerName,
		"duration for apply new config": time.Since(start),
	}).Info("remove all services done")
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
