package application

import (
	"github.com/khannz/crispy-palm-tree/lbost1a-dummy/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-dummy/usecase"
	"github.com/sirupsen/logrus"
)

// DummyFacade struct
type DummyFacade struct {
	DummyWorker domain.DummyWorker
	IDgenerator domain.IDgenerator
	Logging     *logrus.Logger
}

// NewDummyFacade ...
func NewDummyFacade(dummyWorker domain.DummyWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *DummyFacade {

	return &DummyFacade{
		DummyWorker: dummyWorker,
		IDgenerator: idGenerator,
		Logging:     logging,
	}
}

func (dummyFacade *DummyFacade) AddToDummy(ip string) error {
	newAddToDummyEntity := usecase.NewAddToDummyEntity(dummyFacade.DummyWorker)
	return newAddToDummyEntity.AddToDummy(ip)
}

func (dummyFacade *DummyFacade) RemoveFromDummy(ip string) error {
	newAddToDummyEntity := usecase.NewRemoveFromDummyEntity(dummyFacade.DummyWorker)
	return newAddToDummyEntity.RemoveFromDummy(ip)
}
