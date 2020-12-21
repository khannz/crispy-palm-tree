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

			"orch address": viperConfig.GetString(orchAddressName),
			"orch timeout": viperConfig.GetDuration(orchTimeoutName),

			"id type":         viperConfig.GetString(idTypeName),
			"hc address":      viperConfig.GetString(healthcheckAddressName),
			"hc timeout":      viperConfig.GetDuration(responseTimerName),
			"route address":   viperConfig.GetString(routeAddressName),
			"route timeout":   viperConfig.GetDuration(routeTimeoutName),
			"tunnel address":  viperConfig.GetString(tunnelAddressName),
			"tunnel timeout":  viperConfig.GetDuration(tunnelTimeoutName),
			"ip rule address": viperConfig.GetString(ipRuleAddressName),
			"ip rule timeout": viperConfig.GetDuration(ipRuleTimeoutName),
			"dummy address":   viperConfig.GetString(dummyAddressName),
			"dummy timeout":   viperConfig.GetDuration(dummyTimeoutName),
			"ipvs address":    viperConfig.GetString(ipvsAddressName),
			"ipvs timeout":    viperConfig.GetDuration(ipvsTimeoutName),

			"consul address":          viperConfig.GetString(consulAddressName),
			"consul subscribe path":   viperConfig.GetString(consulSubscribePathName),
			"consul app servers path": viperConfig.GetString(consulAppServersPathName),
			"consul service manifest": viperConfig.GetString(consulServiceManifestName),

			"t1 id": viperConfig.GetString(t1OrchIDName),
		}).Info("")

		gracefulShutdown := &domain.GracefulShutdown{}

		// TODO: global locker. consul and get runtime may concurrent. if consul update => retake runtime after apply

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// Workers start
		routeWorker := portadapter.NewRouteWorker(viperConfig.GetString(routeAddressName), viperConfig.GetDuration(routeTimeoutName), logging)
		tunnelWorker := portadapter.NewTunnelWorker(viperConfig.GetString(tunnelAddressName), viperConfig.GetDuration(tunnelTimeoutName), logging)
		ipRuleWorker := portadapter.NewIpRuleWorker(viperConfig.GetString(ipRuleAddressName), viperConfig.GetDuration(ipRuleTimeoutName), logging)

		ipvsWorker := portadapter.NewIpvsWorker(viperConfig.GetString(ipvsAddressName), viperConfig.GetDuration(ipvsTimeoutName), logging)

		dummyWorker := portadapter.NewDummyWorker(viperConfig.GetString(dummyAddressName), viperConfig.GetDuration(dummyTimeoutName), logging)

		// mem init
		memoryWorker := &portadapter.MemoryWorker{
			Services:                     make(map[string]*domain.ServiceInfo),
			ApplicationServersTunnelInfo: make(map[string]int),
		}

		//  healthchecks start
		healthcheckChecker, err := portadapter.NewHealthcheckChecker(viperConfig.GetString(healthcheckAddressName), viperConfig.GetDuration(responseTimerName), logging)
		if err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("connect to healthchecks fail: %v", err)
		}
		defer healthcheckChecker.Conn.Close()

		hc := healthcheck.NewHeathcheckEntity( // memoryWorker,
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

		// TODO: unimplemented read runtime
		grpcServer := application.NewGrpcServer(viperConfig.GetString(orchAddressName), facade, logging) // gorutine inside
		if err := grpcServer.StartServer(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("grpc server start error: %v", err)
		}
		defer grpcServer.CloseServer()

		// consul worker start
		consulWorker, err := application.NewConsulWorker(facade,
			viperConfig.GetString(consulAddressName),
			viperConfig.GetString(consulSubscribePathName),
			viperConfig.GetString(consulAppServersPathName),
			viperConfig.GetString(consulServiceManifestName),
			logging)
		if err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("connect to consul fail: %v", err)
		}
		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("connected to consul")
		go consulWorker.ConsulConfigWatch()
		go consulWorker.JobWorker()

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
