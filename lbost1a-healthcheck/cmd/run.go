package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/application"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "lbost1ah 😉",
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

			"hc address":    viperConfig.GetString(hcAddressName),
			"hc timeout":    viperConfig.GetDuration(hcTimeoutName),
			"ipvs address":  viperConfig.GetString(ipvsAddressName),
			"ipvs timeout":  viperConfig.GetDuration(ipvsTimeoutName),
			"dummy address": viperConfig.GetString(dummyAddressName),
			"dummy timeout": viperConfig.GetDuration(dummyTimeoutName),

			"tech interface": viperConfig.GetString(techInterfaceName),
			"id type":        viperConfig.GetString(idTypeName),
		}).Info("")

		locker := &domain.Locker{}

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// ipvsadmSender init
		ipvsadmSender := portadapter.NewIPVSWorkerEntity(viperConfig.GetString(ipvsAddressName), viperConfig.GetDuration(ipvsTimeoutName), logging)
		if err := ipvsadmSender.ConnectToIPVS(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("IPVSADM can't connect: %v", err)
		}
		defer ipvsadmSender.DisconnectFromIPVS()
		if err := ipvsadmSender.IPVSFlush(idForRootProcess); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("IPVSADM can't flush data at start: %v", err)
		}
		// ipvsadmSender end

		// dummySender init
		dw := portadapter.NewDummyWorkerEntity(viperConfig.GetString(dummyAddressName), viperConfig.GetDuration(dummyTimeoutName), logging)
		if err := dw.ConnectToDummy(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("DUMMY can't connect: %v", err)
		}
		defer dw.DisconnectFromDummy()
		// dummySender end

		// db init
		memDB, err := portadapter.NewStorageEntity(true, "", logging)
		if err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("db init error: %v", err)
		}
		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Debug("init db successful")
		defer memDB.Db.Close()

		// FIXME: client connects

		hc := portadapter.NewHeathcheckEntity(memDB,
			ipvsadmSender,
			viperConfig.GetString(techInterfaceName),
			locker,
			false,
			dw,
			idGenerator,
			logging)

		// init config end

		facade := application.NewHCFacade(locker,
			memDB,
			hc,
			idGenerator,
			logging)

		// up grpc api

		// grpcServer := application.NewGrpcServer(viperConfig.GetString(hcAddressName), facade, logging) // gorutine inside
		grpcServer := application.NewGrpcServer(viperConfig.GetString(hcAddressName), facade, logging) // gorutine inside
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
