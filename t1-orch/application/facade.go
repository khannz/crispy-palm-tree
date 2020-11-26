package application

import (
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/sirupsen/logrus"
)

// T1OrchFacade struct
type T1OrchFacade struct {
	MemoryWorker     domain.MemoryWorker
	RouteMaker       domain.RouteMaker
	HeathcheckEntity *healthcheck.HeathcheckEntity
	GracefulShutdown *domain.GracefulShutdown
	IDgenerator      domain.IDgenerator
	Logging          *logrus.Logger
}

// NewT1OrchFacade ...
func NewT1OrchFacade(MemoryWorker domain.MemoryWorker,
	RouteMaker domain.RouteMaker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *T1OrchFacade {

	return &T1OrchFacade{
		MemoryWorker:     MemoryWorker,
		RouteMaker:       RouteMaker,
		HeathcheckEntity: hc,
		GracefulShutdown: gracefulShutdown,
		IDgenerator:      idGenerator,
		Logging:          logging,
	}
}
