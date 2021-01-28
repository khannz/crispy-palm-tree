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
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lbost1ad",
	Short: "dummy customizer ;-)",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "lbos orchestrator 😉",
	Run: func(cmd *cobra.Command, args []string) {
		idGenerator := chooseIDGenerator()
		idForRootProcess := idGenerator.NewID()

		// validate fields
		logging.WithFields(logrus.Fields{
			"version":          version,
			"build time":       buildTime,
			"event id":         idForRootProcess,
			"config file path": viper.GetString("config-file-path"),
			"log format":       viper.GetString("log-format"),
			"log level":        viper.GetString("log-level"),
			"log output":       viper.GetString("log-output"),
			"syslog tag":       viper.GetString("syslog-tag"),

			"t1 orch id":            viper.GetString("t1-id"),
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

			"consul address":          viper.GetString("consul-address"),
			"consul subscribe path":   viper.GetString("consul-subscribe-path"),
			"consul app servers path": viper.GetString("consul-app-servers-path"),
			"consul service manifest": viper.GetString("consul-manifest-name"),

			"t1 id": viper.GetString("t1-id"),
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
		routeWorker := portadapter.NewRouteWorker(viper.GetString("route-addr"), viper.GetDuration("route-timeout"), logging)
		tunnelWorker := portadapter.NewTunnelWorker(viper.GetString("tunnel-addr"), viper.GetDuration("tunnel-timeout"), logging)
		ipRuleWorker := portadapter.NewIpRuleWorker(viper.GetString("ip-rule-addr"), viper.GetDuration("ip-rule-timeout"), logging)

		ipvsWorker := portadapter.NewIpvsWorker(viper.GetString("ipvs-addr"), viper.GetDuration("ipvs-timeout"), logging)

		dummyWorker := portadapter.NewDummyWorker(viper.GetString("dummy-addr"), viper.GetDuration("dummy-timeout"), logging)

		// mem init
		memoryWorker := &portadapter.MemoryWorker{
			Services:                     make(map[string]*domain.ServiceInfo),
			ApplicationServersTunnelInfo: make(map[string]int),
		}

		//  healthchecks start
		healthcheckChecker, err := portadapter.NewHealthcheckChecker(viper.GetString("hc-address"), viper.GetDuration("hc-timeout"), logging)
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
		grpcServer := application.NewGrpcServer(viper.GetString("orch-addr"), facade, logging) // gorutine inside
		if err := grpcServer.StartServer(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("grpc server start error: %v", err)
		}
		defer grpcServer.CloseServer()

		// consul worker start
		consulWorker, err := application.NewConsulWorker(facade,
			viper.GetString("consul-address"),
			viper.GetString("consul-subscribe-path"),
			viper.GetString("consul-app-servers-path"),
			viper.GetString("consul-manifest-name"),
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
