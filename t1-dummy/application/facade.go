package application

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-dummy/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-dummy/usecase"
	"github.com/sirupsen/logrus"
)

const sendRuntimeConfigName = "send runtime config"

// DummyFacade struct
type DummyFacade struct {
	DummyWorker       domain.DummyWorker
	HealthcheckWorker domain.HealthcheckWorker
	IDgenerator       domain.IDgenerator
	Logging           *logrus.Logger
}

// NewDummyFacade ...
func NewDummyFacade(dummyWorker domain.DummyWorker,
	healthcheckWorker domain.HealthcheckWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *DummyFacade {

	return &DummyFacade{
		DummyWorker:       dummyWorker,
		HealthcheckWorker: healthcheckWorker,
		IDgenerator:       idGenerator,
		Logging:           logging,
	}
}

func (dummyFacade *DummyFacade) AddToDummy(ip string, id string) error {
	newAddToDummyEntity := usecase.NewAddToDummyEntity(dummyFacade.DummyWorker)
	return newAddToDummyEntity.AddToDummy(ip, id)
}

func (dummyFacade *DummyFacade) RemoveFromDummy(ip string, id string) error {
	newAddToDummyEntity := usecase.NewRemoveFromDummyEntity(dummyFacade.DummyWorker)
	return newAddToDummyEntity.RemoveFromDummy(ip, id)
}

func (dummyFacade *DummyFacade) GetDummyRuntimeConfig(id string) (map[string]struct{}, error) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(dummyFacade.DummyWorker)
	return newGetRuntimeConfigEntity.GetDummyRuntimeConfig(id)
}

func (dummyFacade *DummyFacade) TryToSendRuntimeConfig(id string) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(dummyFacade.DummyWorker)
	newHealthcheckSenderEntity := usecase.NewHealthcheckSenderEntity(dummyFacade.HealthcheckWorker)
	for {
		currentConfig, err := newGetRuntimeConfigEntity.GetDummyRuntimeConfig(id)
		if err != nil {
			dummyFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Errorf("failed to get runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := newHealthcheckSenderEntity.SendToHC(currentConfig, id); err != nil {
			dummyFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Warnf("failed to send runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		dummyFacade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": id,
		}).Info("send runtime config to hc success")
		break
	}
}
