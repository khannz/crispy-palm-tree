package run

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	logger "github.com/thevan4/logrus-wrapper"
)

var cfgFile string

// For builds with ldflags
var (
	version   = "unknown"
	buildTime = "unknown"
	// 	commit  = "TBD @ ldflags"
	// 	branch  = "TBD @ ldflags"
)

var (
	logging *logrus.Logger
)

func init() {
	cobra.OnInitialize(initConfig)

	pflag.StringVarP(&cfgFile,
		"config-file-path",
		"c",
		"./lbost1aipr.yaml",
		"Path to config file. Example value: './nw-lb.yaml'")
	pflag.String("log-output",
		"stdout",
		"Log output. Example values: 'stdout', 'syslog'")
	pflag.String("log-level",
		"info",
		"Log level. Example values: 'info', 'debug', 'trace'")
	pflag.String("log-format",
		"text",
		"Log format. Example values: 'text', 'json'")
	pflag.String("syslog-tag",
		"",
		"Syslog tag. Example: 'trac-dgen'")
	pflag.Bool("log-event-location",
		true,
		"Log event location (like python)")

	pflag.String("orch-address",
		"/var/run/lbost1ao.sock",
		"hc address")
	pflag.Duration("orch-timeout",
		2*time.Second,
		"hc timeout")

	pflag.String("ipRule-address",
		"/var/run/lbost1aipr.sock",
		"ipRule address")
	pflag.Duration("ipRule-timeout",
		2*time.Second,
		"ipRule address")

	pflag.String("id-type",
		"nanoid",
		"ID type(nanoid|uuid4)")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	initLogger()
	validateValues()

	rootCmd.AddCommand(runCmd)

}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("can't read config from file, error:", err)
	}
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

func validateValues() {
	switch viper.GetString("id-type") {
	case "nanoid":
	case "id4":
	default:
		logging.Fatalf("unsuported id type: %v; supported types: nanoid|id4", viper.GetString("id-type"))
	}
}
