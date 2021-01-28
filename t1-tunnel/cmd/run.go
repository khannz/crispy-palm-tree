package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/application"
	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lbost1at",
	Short: "tunnel customizer ;-)",
}

var runCmd = &cobra.Command{
	Use: "run",
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

			"orch address":   viper.GetString("orch-address"),
			"orch timeout":   viper.GetDuration("orch-timeout"),
			"tunnel address": viper.GetString("tunnel-address"),
			"tunnel timeout": viper.GetDuration("tunnel-timeout"),
			"id type":        viper.GetString("id-type"),
		}).Info("")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// tunnelConfigurator start
		tunnelConfigurator := portadapter.NewTunnelEntity(logging)
		// tunnelConfigurator end

		// OrchestratorWorker start
		hw := portadapter.NewOrchestratorWorkerEntity(viper.GetString("orch-address"),
			viper.GetDuration("orch-timeout"),
			logging)
		// OrchestratorWorker end

		// init config end

		facade := application.NewTunnelFacade(tunnelConfigurator,
			hw,
			idGenerator,
			logging)

		// try to sendruntime config
		idForSendRuntimeConfig := idGenerator.NewID()
		go facade.TryToSendRuntimeConfig(idForSendRuntimeConfig)

		// up grpc api
		grpcServer := application.NewGrpcServer(viper.GetString("tunnel-address"), facade, logging) // gorutine inside
		if err := grpcServer.StartServer(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("grpc server start error: %v", err)
		}
		defer grpcServer.CloseServer()

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("program running")

		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("got shutdown signal")

		defer logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("program stopped")
	},
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
