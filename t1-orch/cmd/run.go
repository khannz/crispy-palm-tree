package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/application"
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/khannz/crispy-palm-tree/t1-orch/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "lbos orchestrator 😉",
	Run: func(cmd *cobra.Command, args []string) {
		idGenerator := chooseIDGenerator(viperConfig.GetString(idTypeName))
		idForRootProcess := idGenerator.NewID()

		// validate fields
		logging.WithFields(logrus.Fields{
			"version":          version,
			"build time":       buildTime,
			"event id":         idForRootProcess,
			"config file path": viperConfig.GetString(configFilePathName),
			"log format":       viperConfig.GetString(logFormatName),
			"log level":        viperConfig.GetString(logLevelName),
			"log output":       viperConfig.GetString(logOutputName),
			"syslog tag":       viperConfig.GetString(syslogTagName),

			"t1 orch id":            viperConfig.GetString(t1OrchIDName),
			"healthcheck interface": viperConfig.GetString(hlckInterfaceName),

			"id type":       viperConfig.GetString(idTypeName),
			"hc address":    viperConfig.GetString(hcAddressName),
			"hc timeout":    viperConfig.GetDuration(hcTimeoutName),
			"route address": viperConfig.GetString(routeAddressName),
			"route timeout": viperConfig.GetDuration(routeTimeoutName),
			"dummy address": viperConfig.GetString(dummyAddressName),
			"dummy timeout": viperConfig.GetDuration(dummyTimeoutName),
			"ipvs address":  viperConfig.GetString(ipvsAddressName),
			"ipvs timeout":  viperConfig.GetDuration(ipvsTimeoutName),
		}).Info("")

		gracefulShutdown := &domain.GracefulShutdown{}

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// Workers start
		routeWorker := portadapter.NewRouteWorker(viperConfig.GetString(routeAddressName), viperConfig.GetDuration(routeTimeoutName), logging)

		ipvsWorker := portadapter.NewIpvsWorker(viperConfig.GetString(ipvsAddressName), viperConfig.GetDuration(ipvsTimeoutName), logging)

		dummyWorker := portadapter.NewDummyWorker(viperConfig.GetString(dummyAddressName), viperConfig.GetDuration(dummyTimeoutName), logging)

		// mem init
		memoryWorker := &portadapter.MemoryWorker{
			Services:                     make(map[string]*domain.ServiceInfo),
			ApplicationServersTunnelInfo: make(map[string]int),
		}

		//  healthchecks start
		healthcheckChecker := portadapter.NewHealthcheckChecker(viperConfig.GetString(hcAddressName), viperConfig.GetDuration(hcTimeoutName), logging)

		hc := healthcheck.NewHeathcheckEntity(memoryWorker,
			healthcheckChecker,
			ipvsWorker,
			dummyWorker,
			idGenerator,
			logging)

		// healthchecks end

		// init config end

		facade := application.NewT1OrchFacade(memoryWorker,
			routeWorker,
			hc,
			gracefulShutdown,
			idGenerator,
			logging)

		if err := facade.InitConfigAtStart("FIXME: agent id here", idForRootProcess); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("initialize runtime settings fail: %v", err)
		}
		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("initialize runtime settings successful")

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("orch is running")

		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("got shutdown signal")

		gracefulShutdownUsecases(gracefulShutdown, viperConfig.GetDuration(maxShutdownTimeName), logging)

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("rest API is Done")

		// gracefulShutdown.ShutdownNow = false // TODO: so dirty trick
		// if err := facade.DisableRuntimeSettings(viperConfig.GetBool(mockMode), idForRootProcess); err != nil {
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

	ticker := time.NewTicker(time.Duration(100 * time.Millisecond)) // hardcode
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTimeForJobsIsDone)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			gracefulShutdown.Lock()
			if gracefulShutdown.UsecasesJobs <= 0 {
				logging.Info("All jobs is done")
				defer gracefulShutdown.Unlock()
				return
			}
			gracefulShutdown.Unlock()
			continue
		case <-ctx.Done():
			gracefulShutdown.Lock()
			logging.Warnf("%v jobs is fail when program stop", gracefulShutdown.UsecasesJobs)
			defer gracefulShutdown.Unlock()
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

func chooseIDGenerator(idType string) domain.IDgenerator {
	switch viperConfig.GetString(idTypeName) {
	case "nanoid":
		return portadapter.NewIDGenerator()
	case "uuid4":
		return portadapter.NewUUIIDGenerator()
	default:
		return portadapter.NewIDGenerator()
	}
}
