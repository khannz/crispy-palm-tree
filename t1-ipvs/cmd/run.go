package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/application"
	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/portadapter"
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
		logger.Info().
			Str("version", version).
			Str("build_time", buildTime).
			Str("commit", commit).
			Str("branch", branch).
			Str("orch-address", viper.GetString("orch-address")).
			Str("ipvs-address", viper.GetString("ipvs-address")).
			Dur("orch-timeout", viper.GetDuration("orch-timeout")).
			Dur("ipvs-timeout", viper.GetDuration("ipvs-timeout")).
			Msg("runtime config")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)

		// ipvsadmConfigurator start
		ipvsadmConfigurator := portadapter.NewEntity(logger)
		//if err != nil {
		//	logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("can't create IPVSADM entity: %v", err)
		//}
		// ipvsadmConfigurator end

		// orchestratorWorker start
		hw := portadapter.NewOrchestratorWorkerEntity(
			viper.GetString("orch-address"),
			viper.GetDuration("orch-timeout"),
			logger,
		)
		// orchestratorWorker end

		// init config end
		facade := application.NewIPVSFacade(
			ipvsadmConfigurator,
			hw,
			logger,
		)

		// try to sendruntime config FIXME garbage
		//idForSendRuntimeConfig := idGenerator.NewID()
		//go facade.TryToSendRuntimeConfig(idForSendRuntimeConfig)

		// up grpc api
		grpcServer := application.NewGrpcServer(viper.GetString("ipvs-address"), facade, logger) // goroutine inside
		if err := grpcServer.StartServer(); err != nil {
			logger.Fatal().Err(err).Msg("grpc server start error")
		}
		defer grpcServer.CloseServer()

		logger.Info().Msg("application started")

		<-signalChan // shutdown signal

		logger.Warn().Msg("shutdown signal received")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
