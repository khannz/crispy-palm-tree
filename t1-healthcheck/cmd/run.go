package run

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/application"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lbost1ah",
	Short: "healthcheck checker ;-)",
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
			"pprof address":    viper.GetString("pprof-address"),

			"hc address": viper.GetString("hc-address"),

			"id type": viper.GetString("id-type"),
		}).Info("")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// create hc entites
		httpAdvE := portadapter.NewHttpAdvancedEntity(logging)
		httpsE := portadapter.NewHttpsAndHttpsEntity(logging)
		icmpE := portadapter.NewIcmpEntity(logging)
		tcpE := portadapter.NewTcpEntity(logging)
		// ecnd create

		// init config end

		facade := application.NewHCFacade(httpAdvE,
			httpsE,
			icmpE,
			tcpE,
			idGenerator,
			logging)

		// up grpc api
		runtime.SetBlockProfileRate(1)
		grpcServer := application.NewGrpcServer(viper.GetString("hc-address"), viper.GetString("pprof-address"), facade, logging) // gorutine inside
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
