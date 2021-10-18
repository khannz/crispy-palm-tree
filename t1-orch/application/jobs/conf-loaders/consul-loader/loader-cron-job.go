package consul_loader

import (
	"context"
	"sync"
	"sync/atomic"

	appJobs "github.com/gradusp/go-platform/app/jobs"
	"github.com/gradusp/go-platform/pkg/backoff"
	"github.com/gradusp/go-platform/pkg/patterns/observer"
	"github.com/gradusp/go-platform/pkg/scheduler"
	"github.com/gradusp/go-platform/pkg/tm"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
	consulProvider "github.com/khannz/crispy-palm-tree/t1-orch/providers/consul"
	"github.com/pkg/errors"
)

//Config config to create job
type Config struct {
	ConsulProviderConfig consulProvider.ProviderConfig
	JobName              string
	TaskScheduler        scheduler.Scheduler
	JobBackoff           backoff.Backoff
	JobExecutor          tm.TaskManger
}

//NewCronJob new cron job for load data from Consul
func NewCronJob(ctx context.Context, conf Config) (appJobs.JobScheduler, error) {
	const api = "consul-loader/NewCronJob"

	provider, err := consulProvider.NewServiceConfigProvider(conf.ConsulProviderConfig)
	if err != nil {
		return nil, errors.Wrap(err, api)
	}
	defer func() {
		if provider != nil {
			_ = provider.Close()
		}
	}()
	loader := &loaderFromConsul{
		provider: provider,
	}
	jobConf := appJobs.JobSchedulerConf{
		JobID:         conf.JobName,
		NewTask:       loader.produceJob,
		Backoff:       conf.JobBackoff,
		TaskScheduler: conf.TaskScheduler,
	}
	if t := conf.JobExecutor; t != nil {
		jobConf.TaskManager = func(_ context.Context) tm.TaskManger {
			return t
		}
	}
	if loader.JobScheduler, err = appJobs.NewJobScheduler(ctx, jobConf); err != nil {
		return nil, errors.Wrap(err, api)
	}
	subj := loader.JobScheduler.Subject()
	obs := observer.NewObserver(loader.jobEventsObserver, false, appJobs.OnJobFinished{})
	subj.ObserversAttach(obs)
	loader.closers = append(loader.closers, func() error {
		subj.ObserversDetach(obs)
		return nil
	})
	loader.closers = append(loader.closers, loader.JobScheduler.Close)
	loader.closers = append(loader.closers, loader.provider.Close)
	provider = nil
	return loader, nil
}

//-------------------------------------------------------  IMPL -------------------------------------------------//

type loaderFromConsul struct {
	appJobs.JobScheduler
	waitIndex uint64
	closeOnce sync.Once
	provider  providers.ServicesConfigProvider

	closers []func() error
}

//Close impl io.Closer
func (loader *loaderFromConsul) Close() error {
	loader.closeOnce.Do(func() {
		closers := loader.closers
		loader.closers = nil
		for _, c := range closers {
			_ = c()
		}
	})
	return nil
}

func (loader *loaderFromConsul) produceJob(ctx context.Context) (tm.Task, []interface{}, error) {
	task, err := tm.MakeSimpleTask("1", loader.provider.Get)
	if err != nil {
		return nil, nil, err
	}
	args := []interface{}{ctx}
	if idx := atomic.LoadUint64(&loader.waitIndex); idx != 0 {
		args = append(args, consulProvider.WithWaitIndex{WaitIndex: idx})
	}
	return task, args, nil
}

func (loader *loaderFromConsul) jobEventsObserver(event observer.EventType) {
	switch evt := event.(type) {
	case appJobs.OnJobFinished:
		out, ok := evt.JobResult.(appJobs.JobOutput)
		if !ok {
			return
		}
		for _, o := range out.Output {
			if jobOutput, _ := o.(providers.ServicesConfig); jobOutput != nil {
				if p1, ok := jobOutput.(consulProvider.ServiceTransportData); ok {
					atomic.StoreUint64(&loader.waitIndex, p1.QueryMeta.LastIndex)
					return
				}
			}
		}
	}
}
