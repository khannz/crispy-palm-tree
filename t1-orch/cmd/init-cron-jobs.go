package run

import (
	"context"
	"sync"
	"time"

	appJobs "github.com/gradusp/go-platform/app/jobs"
	"github.com/gradusp/go-platform/pkg/backoff"
	pkgNet "github.com/gradusp/go-platform/pkg/net"
	"github.com/gradusp/go-platform/pkg/patterns/observer"
	"github.com/gradusp/go-platform/pkg/scheduler"
	"github.com/gradusp/go-platform/pkg/tm"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	consulJobLoader "github.com/khannz/crispy-palm-tree/t1-orch/application/jobs/conf-loaders/consul-loader"
	waddleJobLoader "github.com/khannz/crispy-palm-tree/t1-orch/application/jobs/conf-loaders/waddle-loader"
	waddleService "github.com/khannz/crispy-palm-tree/t1-orch/external-api/waddle"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
	consulProvider "github.com/khannz/crispy-palm-tree/t1-orch/providers/consul"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers/waddle"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type jobConstructor struct {
	ctx        context.Context
	taskManger tm.TaskManger
	scheduler  scheduler.Scheduler
	backoff    backoff.Backoff
	observers  []observer.Observer
}

func (constructor *jobConstructor) constructConsulJob() (appJobs.JobScheduler, error) {
	conf := consulJobLoader.Config{
		ConsulProviderConfig: consulProvider.ProviderConfig{
			ConsulAddress:        viper.GetString("consul-address"),
			ConsulSubscribePath:  viper.GetString("consul-subscribe-path"),
			ConsulAppServersPath: viper.GetString("consul-app-servers-path"),
			ServiceManifest:      viper.GetString("consul-manifest-name"),
		},
		JobName:       "consul-job-loader",
		JobExecutor:   constructor.taskManger,
		TaskScheduler: constructor.scheduler,
		JobBackoff:    constructor.backoff,
	}
	ctx, cancel := context.WithCancel(constructor.ctx)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	sourceLoaderJob, err := consulJobLoader.NewCronJob(ctx, conf)
	if err != nil {
		return nil, err
	}
	ret := &wrappedJob{
		JobScheduler: sourceLoaderJob,
	}
	subj := ret.Subject()
	subj.ObserversAttach(constructor.observers...)
	c := cancel
	cancel = nil
	ret.closers = append(ret.closers,
		func() error {
			subj.ObserversDetach(constructor.observers...)
			return nil
		},
		func() error {
			c()
			return nil
		},
		sourceLoaderJob.Close,
	)
	return ret, nil
}

func (constructor *jobConstructor) constructWaddleJob() (appJobs.JobScheduler, error) {
	waddleAddr := viper.GetString("waddle-address")
	nodeIP := viper.GetString("waddle-node-ip")

	ctx, cancel := context.WithCancel(constructor.ctx)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	endpoint, err := pkgNet.ParseEndpoint(waddleAddr)
	if err != nil {
		return nil, err
	}
	if endpoint.IsUnixDomain() {
		waddleAddr = endpoint.FQN()
	} else {
		waddleAddr = endpoint.String()
	}
	opts := []grpcRetry.CallOption{
		//grpcRetry.WithBackoff(grpcRetry.BackoffExponential(5 * time.Millisecond)),
		grpcRetry.WithBackoff(grpcRetry.BackoffLinear(500 * time.Millisecond)),
		grpcRetry.WithMax(10),
	}
	var gConn *grpc.ClientConn
	gConn, err = grpc.DialContext(ctx,
		waddleAddr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(grpcRetry.UnaryClientInterceptor(opts...)),
	)
	if err != nil {
		return nil, err
	}
	var client waddleService.NodeInfoServiceClient
	client, err = waddleService.NewNodeInfoServiceClient(ctx, gConn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()
	var dataProv providers.ServicesConfigProvider
	dataProv, err = waddle.NewServiceConfigProvider(waddle.ProviderConfig{
		WaddleServiceClient: client,
	})
	if err != nil {
		return nil, err
	}
	client = nil
	defer func() {
		if dataProv != nil {
			_ = dataProv.Close()
		}
	}()
	jobConf := waddleJobLoader.Config{
		JobName:       "waddle-job-loader",
		NodeIP:        nodeIP,
		Provider:      dataProv,
		JobExecutor:   constructor.taskManger,
		TaskScheduler: constructor.scheduler,
		JobBackoff:    constructor.backoff,
	}
	var sourceLoaderJob appJobs.JobScheduler
	sourceLoaderJob, err = waddleJobLoader.NewCronJob(ctx, jobConf)
	if err != nil {
		return nil, err
	}

	ret := &wrappedJob{
		JobScheduler: sourceLoaderJob,
	}
	subj := ret.Subject()
	subj.ObserversAttach(constructor.observers...)
	c := cancel
	cancel = nil
	ret.closers = append(ret.closers,
		func() error {
			subj.ObserversDetach(constructor.observers...)
			return nil
		},
		func() error {
			c()
			return nil
		},
		sourceLoaderJob.Close,
		dataProv.Close,
	)
	dataProv = nil
	return ret, nil
}

type wrappedJob struct {
	appJobs.JobScheduler
	closeOnce sync.Once
	closers   []func() error
}

//Close io.Closer
func (job *wrappedJob) Close() error {
	job.closeOnce.Do(func() {
		closers := job.closers
		job.closers = nil
		for _, c := range closers {
			_ = c()
		}
	})
	return nil
}
