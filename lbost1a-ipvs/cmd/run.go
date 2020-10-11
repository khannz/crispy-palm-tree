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
)

var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "lbost1ai 😉",
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

			"hc address":     viperConfig.GetString(ipvsAddressName),
			"tech interface": viperConfig.GetString(techInterfaceName),
			"id type":        viperConfig.GetString(idTypeName),
		}).Info("")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// ipvsadmConfigurator start
		ipvsadmConfigurator, err := portadapter.NewIPVSADMEntity(logging)
		if err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("can't create IPVSADM entity: %v", err)
		}
		if err := ipvsadmConfigurator.IPVSFlush(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("IPVSADM can't IPVSFlush data at start: %v", err)
		}
		// ipvsadmConfigurator end

		// init config end

		facade := application.NewIPVSFacade(ipvsadmConfigurator,
			idGenerator,
			logging)

		// up grpc api

		// grpcServer := application.NewGrpcServer(viperConfig.GetString(ipvsAddressName), facade, logging) // gorutine inside
		grpcServer := application.NewGrpcServer(viperConfig.GetString(ipvsAddressName), facade, logging) // gorutine inside
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
