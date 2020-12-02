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
		Logging:          logging,
	}
}

func (t1OrchFacade *T1OrchFacade) ApplyNewConfig(updatedServicesInfo []*domain.ServiceInfo) error {
	id := t1OrchFacade.IDgenerator.NewID()

	newNewServiceEntity := usecase.NewNewServiceEntity(t1OrchFacade.MemoryWorker, t1OrchFacade.RouteWorker, t1OrchFacade.HeathcheckEntity, t1OrchFacade.GracefulShutdown, t1OrchFacade.Logging)
	for _, currentService := range updatedServicesInfo {
		if err := t1OrchFacade.MemoryWorker.AddService(currentService); err != nil {
			return err
		}
		if err := newNewServiceEntity.NewService(currentService, id); err != nil {
			return err
		}
	}
	return nil
}
