package waddle_loader

import (
	"context"
	"time"

	appJobs "github.com/gradusp/go-platform/app/jobs"
	"github.com/gradusp/go-platform/pkg/backoff"
	"github.com/gradusp/go-platform/pkg/scheduler"
	"github.com/gradusp/go-platform/pkg/tm"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers/waddle"
	"github.com/pkg/errors"
)

//Config config to create job
type Config struct {
	Provider       providers.ServicesConfigProvider //is not owned
	JobName        string
	JobMaxDuration time.Duration
	NodeIP         string
	TaskScheduler  scheduler.Scheduler //is not owned
	JobBackoff     backoff.Backoff
	JobExecutor    tm.TaskManger //is not owned
}

type loader = Config

//NewCronJob ...
func NewCronJob(ctx context.Context, conf Config) (appJobs.JobScheduler, error) {
	const api = "waddle-loader/NewCronJob"

	jobConf := appJobs.JobSchedulerConf{
		JobID:         conf.JobName,
		NewTask:       conf.produceJob,
		Backoff:       conf.JobBackoff,
		TaskScheduler: conf.TaskScheduler,
	}
	if t := conf.JobExecutor; t != nil {
		jobConf.TaskManager = func(_ context.Context) tm.TaskManger {
			return t
		}
	}
	sched, err := appJobs.NewJobScheduler(ctx, jobConf)
	return sched, errors.Wrap(err, api)
}

func (loader *loader) produceJob(ctx context.Context) (tm.Task, []interface{}, error) {
	f := loader.Provider.Get
	if loader.JobMaxDuration > 0 {
		f = func(ctx context.Context, opts ...providers.GetOption) (providers.ServicesConfig, error) {
			ctx1, c := context.WithTimeout(ctx, loader.JobMaxDuration)
			defer c()
			return loader.Provider.Get(ctx1, opts...)
		}
	}
	task, err := tm.MakeSimpleTask("1", f)
	if err != nil {
		return nil, nil, err
	}
	args := []interface{}{ctx}
	if len(loader.NodeIP) > 0 {
		args = append(args, waddle.WithNodeIP{NodeIp: loader.NodeIP})
	}
	return task, args, nil
}
