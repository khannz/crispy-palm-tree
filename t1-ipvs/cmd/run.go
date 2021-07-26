package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/application"
	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lbost1ai",
	Short: "ipvs tuner",
}

var runCmd = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {
		logging.WithFields(logrus.Fields{
			"version":          version,
			"build time":       buildTime,
			"event id":         idForRootProcess,
			"config file path": viper.GetString("config-file-path"),
			"log format":       viper.GetString("log-format"),
			"log level":        viper.GetString("log-level"),
			"log output":       viper.GetString("log-output"),
			"syslog tag":       viper.GetString("syslog-tag"),

			"hc address":   viper.GetString("orch-address"),
			"hc timeout":   viper.GetDuration("orch-timeout"),
			"ipvs address": viper.GetString("ipvs-address"),
			"id type":      viper.GetString("id-type"),
		}).Info("")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)

		// ipvsadmConfigurator start
		ipvsadmConfigurator, err := portadapter.NewIPVSADMEntity(logging)
		// ipvsadmConfigurator end

		// orchestratorWorker start
		hw := portadapter.NewOrchestratorWorkerEntity(
			viper.GetString("orch-address"),
			viper.GetDuration("orch-timeout"),
			logging)
		)
		// orchestratorWorker end

		// init config end
		facade := application.NewIPVSFacade(
			ipvsadmConfigurator,
			hw,
			logging)

		// try to sendruntime config FIXME garbage
		//idForSendRuntimeConfig := idGenerator.NewID()
		//go facade.TryToSendRuntimeConfig(idForSendRuntimeConfig)

		// up grpc api
		grpcServer := application.NewGrpcServer(viper.GetString("ipvs-address"), facade, logging) // gorutine inside
		if err := grpcServer.StartServer(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("grpc server start error: %v", err)
		}
		defer grpcServer.CloseServer()

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("program running")

		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("got shutdown signal")

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("program stopped")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
