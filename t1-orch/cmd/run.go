package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	appJobs "github.com/gradusp/go-platform/app/jobs"
	"github.com/gradusp/go-platform/pkg/backoff"
	"github.com/gradusp/go-platform/pkg/functional"
	"github.com/gradusp/go-platform/pkg/patterns/observer"
	"github.com/gradusp/go-platform/pkg/scheduler"
	"github.com/gradusp/go-platform/pkg/tm"
	"github.com/khannz/crispy-palm-tree/t1-orch/application"
	"github.com/khannz/crispy-palm-tree/t1-orch/application/jobs/consumers"
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/khannz/crispy-palm-tree/t1-orch/portadapter"
	consulProvider "github.com/khannz/crispy-palm-tree/t1-orch/providers/consul"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lbost1ao",
	Short: "lbos orchestrator ;-)",
}

var runCmd = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {
		idGenerator := chooseIDGenerator()
		idForRootProcess := idGenerator.NewID()

		// validate fields

		fields := logrus.Fields{
			"version":          version,
			"build time":       buildTime,
			"event id":         idForRootProcess,
			"config file path": viper.GetString("config-file-path"),
			"log format":       viper.GetString("log-format"),
			"log level":        viper.GetString("log-level"),
			"log output":       viper.GetString("log-output"),
			"syslog tag":       viper.GetString("syslog-tag"),

			"healthcheck interface": viper.GetString("hlck-interface"),

			"orch address": viper.GetString("orch-addr"),
			"orch timeout": viper.GetDuration("orch-timeout"),

			"id type":         viper.GetString("id-type"),
			"hc address":      viper.GetString("hc-address"),
			"hc timeout":      viper.GetDuration("hc-timeout"),
			"route address":   viper.GetString("route-addr"),
			"route timeout":   viper.GetDuration("route-timeout"),
			"tunnel address":  viper.GetString("tunnel-addr"),
			"tunnel timeout":  viper.GetDuration("tunnel-timeout"),
			"ip rule address": viper.GetString("ip-rule-addr"),
			"ip rule timeout": viper.GetDuration("ip-rule-timeout"),
			"dummy address":   viper.GetString("dummy-addr"),
			"dummy timeout":   viper.GetDuration("dummy-timeout"),
			"ipvs address":    viper.GetString("ipvs-addr"),
			"ipvs timeout":    viper.GetDuration("ipvs-timeout"),

			"use data provider": viper.GetString("source-provider"), //"consul"|"waddle"
			//"t1 id": viper.GetString("t1-id"),
		}

		if viper.GetString("source-provider") == ConsulDataProvider {
			fields["consul address"] = viper.GetString("consul-address")
			fields["consul subscribe path"] = viper.GetString("consul-subscribe-path")
			fields["consul app servers path"] = viper.GetString("consul-app-servers-path")
			fields["consul service manifest"] = viper.GetString("consul-manifest-name")
		} else {
			fields["waddle address"] = viper.GetString("waddle-address")
			fields["waddle node IP"] = viper.GetString("waddle-node-ip")
		}

		logging.WithFields(fields).Info("")

		gracefulShutdown := new(domain.GracefulShutdown)

		// TODO: global locker. consul and get runtime may concurrent. if consul update => retake runtime after apply

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// Workers start
		routeWorker := portadapter.NewRouteWorker(viper.GetString("route-addr"), viper.GetDuration("route-timeout"), logging)
		tunnelWorker := portadapter.NewTunnelWorker(viper.GetString("tunnel-addr"), viper.GetDuration("tunnel-timeout"), logging)
		ipRuleWorker := portadapter.NewIpRuleWorker(viper.GetString("ip-rule-addr"), viper.GetDuration("ip-rule-timeout"), logging)

		ipvsWorker := portadapter.NewIpvsWorker(viper.GetString("ipvs-addr"), viper.GetDuration("ipvs-timeout"), logging)

		dummyWorker := portadapter.NewDummyWorker(viper.GetString("dummy-addr"), viper.GetDuration("dummy-timeout"), logging)

		// mem init
		memoryWorker := &portadapter.MemoryWorker{
			Services:                     make(domain.ServiceInfoConf),
			ApplicationServersTunnelInfo: make(map[string]int),
		}

		//  healthchecks start
		healthcheckChecker, err := portadapter.NewHealthcheckChecker(viper.GetString("hc-address"), viper.GetDuration("hc-timeout"), logging)
		if err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("connect to healthchecks fail: %v", err)
		}
		defer healthcheckChecker.Conn.Close()

		hc := healthcheck.NewHealthcheckEntity( // memoryWorker,
			healthcheckChecker,
			ipvsWorker,
			dummyWorker,
			idGenerator,
			logging)

		// healthchecks end

		// init config end

		facade := application.NewT1OrchFacade(memoryWorker,
			tunnelWorker,
			routeWorker,
			ipRuleWorker,
			hc,
			gracefulShutdown,
			idGenerator,
			logging)

		jobScheduleInterval := scheduler.NewConstIntervalScheduler(time.Minute) //TODO: может вынести в конфиг?
		jc := &jobConstructor{
			ctx:        context.Background(),
			taskManger: tm.NewTaskManager(),
			scheduler:  jobScheduleInterval,
			backoff: backoff.ExponentialBackoffBuilder().
				WithInitialInterval(500 * time.Millisecond).
				WithMaxInterval(5 * time.Minute).
				Build(),
		}
		jc.observers = append(jc.observers,
			observer.NewObserver(func(evt observer.EventType) {
				switch t := evt.(type) {
				case appJobs.OnJobFinished:
					e := t.FindError()
					if e == nil {
						break
					}
					isFatal := errors.Is(e, functional.ErrArgsNotMatched2Signature) ||
						errors.Is(e, consulProvider.ErrDataNotFit2Model)

					if isFatal {
						logging.Fatalf("from job '%s' got fataL %v", t.JobID, e)
					}
				case appJobs.OnJobSchedulerStop:
					logging.Fatalf("from job '%s' got fataL %v", t.JobID, t.Reason)
				case appJobs.OnJobLog:
					logging.Infof("%s", t)
				}
			}, false,
				appJobs.OnJobLog{},
				appJobs.OnJobFinished{},
				appJobs.OnJobSchedulerStop{},
			),
		)

		var sourceLoaderJob appJobs.JobScheduler
		if viper.GetString("source-provider") == ConsulDataProvider {
			sourceLoaderJob, err = jc.constructConsulJob()
		} else {
			sourceLoaderJob, err = jc.constructWaddleJob()
		}
		if err != nil {
			logging.Fatal(err.Error())
		}
		consumerCloser := consumers.NewFacadeConsumer(sourceLoaderJob, facade, logging)

		// TODO: unimplemented read runtime
		grpcServer := application.NewGrpcServer(viper.GetString("orch-addr"), facade, logging) // gorutine inside
		if err = grpcServer.StartServer(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("grpc server start error: %v", err)
		}
		sourceLoaderJob.Enable(true)
		sourceLoaderJob.Schedule()

		<-signalChan // shutdown signal

		for _, c := range []io.Closer{consumerCloser, sourceLoaderJob, jobScheduleInterval} {
			_ = c.Close()
		}
		grpcServer.CloseServer()

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("got shutdown signal")

		gracefulShutdownUsecases(gracefulShutdown, viper.GetDuration("max-shutdown-time"), logging)

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("rest API is Done")

		// gracefulShutdown.ShutdownNow = false // TODO: so dirty trick
		// if err := facade.DisableRuntimeSettings(viper.GetBool(mockMode), idForRootProcess); err != nil {
		// 	logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Warnf("disable runtime settings errors: %v", err)
		// }

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("runtime settings disabled")

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("orchestrator stopped")
	},
}

func gracefulShutdownUsecases(gracefulShutdown *domain.GracefulShutdown, maxWaitTimeForJobsIsDone time.Duration, logging *logrus.Logger) {
	gracefulShutdown.Lock()
	gracefulShutdown.ShutdownNow = true
	gracefulShutdown.Unlock()

	ticker := time.NewTicker(100 * time.Millisecond) // hardcode
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTimeForJobsIsDone)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			gracefulShutdown.Lock()
			state := gracefulShutdown.UsecasesJobs
			gracefulShutdown.Unlock()
			if state <= 0 {
				logging.Info("All jobs is done")
				return
			}
		case <-ctx.Done():
			gracefulShutdown.Lock()
			state := gracefulShutdown.UsecasesJobs
			gracefulShutdown.Unlock()
			logging.Warnf("%v jobs is fail when program stop", state)
			return
		}
	}
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func chooseIDGenerator() domain.IDgenerator {
	switch viper.GetString("id-type") {
	case "nanoid":
		return portadapter.NewIDGenerator()
	case "uuid4":
		return portadapter.NewUUIIDGenerator()
	default:
		return portadapter.NewIDGenerator()
	}
}
