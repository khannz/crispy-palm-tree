package consumers

import (
	"errors"
	"io"
	"sync"

	appJobs "github.com/gradusp/go-platform/app/jobs"
	"github.com/gradusp/go-platform/pkg/patterns/observer"
	consulProviderAdapter "github.com/khannz/crispy-palm-tree/t1-orch/application/adapters/consul-provider"
	_ "github.com/khannz/crispy-palm-tree/t1-orch/application/adapters/waddle-provider"
	waddleProviderAdapter "github.com/khannz/crispy-palm-tree/t1-orch/application/adapters/waddle-provider"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
	consulProvider "github.com/khannz/crispy-palm-tree/t1-orch/providers/consul"
	waddleProvider "github.com/khannz/crispy-palm-tree/t1-orch/providers/waddle"
	"github.com/sirupsen/logrus"
)

//NewFacadeConsumer ...
func NewFacadeConsumer(job appJobs.JobScheduler, facade FacadeInterface, logger *logrus.Logger) io.Closer {
	ret := &facadeConsumer{
		JobName: job.ID(),
		facade:  facade,
		logger:  logger,
	}
	subj := job.Subject()
	obs := observer.NewObserver(ret.observe, false, appJobs.OnJobFinished{})
	subj.ObserversAttach(obs)
	ret.close = func() {
		subj.ObserversDetach(obs)
	}
	return ret
}

//facadeConsumer observer jobs results and run facade methods if need
type facadeConsumer struct {
	JobName string
	facade  FacadeInterface
	logger  *logrus.Logger

	closeOnce sync.Once
	close     func()
}

//Close io.Closer
func (cons *facadeConsumer) Close() error {
	cons.closeOnce.Do(func() {
		cons.close()
		cons.close = nil
	})
	return nil
}

func (cons *facadeConsumer) observe(evt observer.EventType) {
	switch fin := evt.(type) {
	case appJobs.OnJobFinished:
		if fin.JobID != cons.JobName {
			return
		}
		switch t := fin.JobResult.(type) {
		case appJobs.JobOutput:
			for _, item := range t.Output {
				if c, ok := item.(providers.ServicesConfig); ok {
					var err error
					switch data := c.(type) {
					case consulProvider.ServiceTransportData:
						err = cons.consumeConsulData(data)
					case waddleProvider.ServiceConfigData:
						err = cons.consumeWaddleData(data)
					}
					if err != nil && cons.logger != nil {
						cons.logger.WithFields(logrus.Fields{
							"entity": cons.JobName,
						}).Errorf("config update error: %v", err)
					}
					return
				}
			}
		}
	}
}

func (cons *facadeConsumer) consumeConsulData(data consulProvider.ServiceTransportData) error {
	converted, err := consulProviderAdapter.ServiceTransportData(data).ToServiceInfoConf()
	if err == nil {
		err = cons.facade.ApplyNewConfig(converted)
	} else if errors.Is(err, consulProviderAdapter.ErrNonePayload) {
		err = cons.facade.RemoveAllConfigs()
	}
	return err
}

func (cons *facadeConsumer) consumeWaddleData(data waddleProvider.ServiceConfigData) error {
	converted, err := waddleProviderAdapter.ServiceConfigData(data).ToServiceInfoConf()
	if err == nil {
		err = cons.facade.ApplyNewConfig(converted)
	}
	return err
}
