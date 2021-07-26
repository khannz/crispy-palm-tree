package run

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	logger "github.com/thevan4/logrus-wrapper"
)

// For builds with ldflags
var (
	logging *logrus.Logger
	version   = "unknown" /* TBD @ ldflags */
	buildTime = "unknown" /* TBD @ ldflags */
	commit    = "unknown" /* TBD @ ldflags */
	branch    = "unknown" /* TBD @ ldflags */
)

func init() {
	cobra.OnInitialize(initConfig)

	//rootCmd.PersistentFlags().String(
	//	"log-output",
	//	"stdout",
	//	"Log output. Example values: 'stdout', 'syslog'",
	//)
	//rootCmd.PersistentFlags().String(
	//	"log-level",
	//	"info",
	//	"Log level. Example values: 'info', 'debug', 'trace'",
	//)
	//rootCmd.PersistentFlags().String(
	//	"log-format",
	//	"text",
	//	"Log format. Example values: 'text', 'json'",
	//)
	rootCmd.PersistentFlags().String(
		"orch-address",
		"/var/run/lbost1ao.sock",
		"hc address",
	)
	rootCmd.PersistentFlags().String(
		"ipvs-address",
		"/var/run/lbost1ai.sock",
		"ipvs address",
	)
	rootCmd.PersistentFlags().Duration(
		"orch-timeout",
		2*time.Second,
		"hc timeout",
	)
	rootCmd.PersistentFlags().Duration(
		"ipvs-timeout",
		2*time.Second,
		"ipvs timeout",
	)

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(runCmd)
}

func initConfig() {
	initEnv()
	initLogger()
}

func initEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func initLogger() {
	newLogger := &logger.Logger{
		Output:           []string{viper.GetString("log-output")},
		Level:            viper.GetString("log-level"),
		Formatter:        viper.GetString("log-format"),
		LogEventLocation: viper.GetBool("log-event-location"),
	}
	var err error
	logging, err = logger.NewLogrusLogger(newLogger)
	if err != nil {
		fmt.Println("init log error:", err)
		os.Exit(1)
	}
}
}
