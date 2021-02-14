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

	rootCmd.PersistentFlags().StringVarP(&cfgFile,
		"config-file-path",
		"c",
		"/opt/lbost1ao/lbost1ao.yaml",
		"Path to config file. Example value: '/opt/lbost1ao/lbost1ao.yaml'")
	rootCmd.PersistentFlags().String("log-output",
		"stdout",
		"Log output. Example values: 'stdout',"+
			" 'syslog'")
	rootCmd.PersistentFlags().String("log-level",
		"info",
		"Log level. Example values: 'info',"+
			" 'debug',"+
			" 'trace'")
	rootCmd.PersistentFlags().String("log-format",
		"text",
		"Log format. Example values: 'text',"+
			" 'json'")
	rootCmd.PersistentFlags().String("syslog-tag",
		"",
		"Syslog tag. Example: 'trac-dgen'")
	rootCmd.PersistentFlags().Bool("log-event-location",
		true,
		"Log event location (like python)")

	rootCmd.PersistentFlags().String("t1-id",
		"bad1",
		"t1 orch id")

	rootCmd.PersistentFlags().String("hlck-interface",
		"",
		"healthcheck interface")

	rootCmd.PersistentFlags().String("id-type",
		"nanoid",
		"ID type(nanoid|uuid4)")

	rootCmd.PersistentFlags().String("orch-addr",
		"/var/run/lbost1ao.sock",
		"orch address. Example:'/var/run/lbost1ao.sock'")
	rootCmd.PersistentFlags().Duration("orch-timeout",
		2*time.Second,
		"orch request timeout")

	rootCmd.PersistentFlags().String("hc-address",
		"/var/run/lbost1ah.sock",
		"Healthcheck address. Example:'127.0.0.1:7000'")
	rootCmd.PersistentFlags().Duration("hc-timeout",
		2*time.Second,
		"Healthcheck request timeout")

	rootCmd.PersistentFlags().String("dummy-addr",
		"/var/run/lbost1ad.sock",
		"dummy address. Example:'/var/run/lbost1ad.sock'")
	rootCmd.PersistentFlags().Duration("dummy-timeout",
		2*time.Second,
		"dummy request timeout")
	rootCmd.PersistentFlags().String("route-addr",
		"/var/run/lbost1ar.sock",
		"route address. Example:'/var/run/lbost1ar.sock'")
	rootCmd.PersistentFlags().Duration("route-timeout",
		2*time.Second,
		"route request timeout")
	rootCmd.PersistentFlags().String("tunnel-addr",
		"/var/run/lbost1at.sock",
		"tunnel address. Example:'/var/run/lbost1at.sock'")
	rootCmd.PersistentFlags().Duration("tunnel-timeout",
		2*time.Second,
		"tunnel request timeout")
	rootCmd.PersistentFlags().String("ip-rule-addr",
		"/var/run/lbost1aipr.sock",
		"ip rule address. Example:'/var/run/lbost1aipr.sock'")
	rootCmd.PersistentFlags().Duration("ip-rule-timeout",
		2*time.Second,
		"ip rule request timeout")
	rootCmd.PersistentFlags().String("ipvs-addr",
		"/var/run/lbost1ai.sock",
		"ipvs address. Example:'/var/run/lbost1ai.sock'")
	rootCmd.PersistentFlags().Duration("ipvs-timeout",
		2*time.Second,
		"ipvs request timeout")

	rootCmd.PersistentFlags().String("consul-address",
		"127.0.0.1:18700",
		"consul address")
	rootCmd.PersistentFlags().String("consul-subscribe-path",
		"lbos/t1-cluster-1/",
		"consul subscribe path")
	rootCmd.PersistentFlags().String("consul-app-servers-path",
		"app-servers/",
		"consul app servers path")
	rootCmd.PersistentFlags().String("consul-manifest-name",
		"manifest",
		"consul service manifest")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(runCmd)
}

func initConfig() {
	initEnv()
	initCfgFile()
	initLogger()
	validateValues()
}

func initEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func initCfgFile() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
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
	if viper.GetString("t1-id") == "bad1" {
		logging.Fatalf("t1 id must be set")
	}

	switch viper.GetString("id-type") {
	case "nanoid":
	case "uuid4":
	default:
		logging.Fatalf("unsuported id type: %v; supported types: nanoid|uuid4", viper.GetString("id-type"))
	}
}
