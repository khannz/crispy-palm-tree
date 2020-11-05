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
)

var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "lbost1at 😉",
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

			"sysctl config path": viperConfig.GetString(sysctlConfigsPathName),
			"ipvs address":       viperConfig.GetString(sockPathName),
			"tech interface":     viperConfig.GetString(sockFilenameName),
			"id type":            viperConfig.GetString(idTypeName),
		}).Info("")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// tunnel maker start
		mockMode := false                                                                            // TODO: remove hardcode
		TunnelWorker := portadapter.NewTunnelFileMaker(viperConfig.GetString(sysctlConfigsPathName), // TODO: refactor. tunnels, routes and sysctl config maker
			mockMode,
			logging)
		// tunnel maker end

		// init config end

		facade := application.NewTunnelFacade(TunnelWorker,
			idGenerator,
			logging)

		// up uds grpc api

		// grpcServer := application.NewGrpcServer(viperConfig.GetString(ipvsAddressName), facade, logging) // gorutine inside
		grpcServer := application.NewUdsGrpcServer(viperConfig.GetString(sockPathName), viperConfig.GetString(sockFilenameName), facade, logging) // gorutine inside
		if err := grpcServer.StartServer(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("grpc server start error: %v", err)
		}
		defer grpcServer.CloseServer()

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("tunner service running")

		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("got shutdown signal")

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("tunner service stopped")
	},
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
