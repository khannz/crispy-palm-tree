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

	// work with flags
	pflag.StringP("config-file-path",
		"c",
		"./lbost1ao.yaml",
		"Path to config file. Example value: './nw-lb.yaml'")
	pflag.String("log-output",
		"stdout",
		"Log output. Example values: 'stdout',"+
			" 'syslog'")
	pflag.String("log-level",
		"info",
		"Log level. Example values: 'info',"+
			" 'debug',"+
			" 'trace'")
	pflag.String("log-format",
		"text",
		"Log format. Example values: 'text',"+
			" 'json'")
	pflag.String("syslog-tag",
		"",
		"Syslog tag. Example: 'trac-dgen'")
	pflag.Bool("log-event-location",
		true,
		"Log event location (like python)")

	pflag.String("t1-id",
		"bad1",
		"t1 orch id")

	pflag.String("hlck-interface",
		"",
		"healthcheck interface")

	pflag.String("id-type",
		"nanoid",
		"ID type(nanoid|uuid4)")

	pflag.String("orch-addr",
		"/var/run/lbost1ao.sock",
		"orch address. Example:'/var/run/lbost1ao.sock'")
	pflag.Duration("orch-timeout",
		2*time.Second,
		"orch request timeout")

	pflag.String("hc-address",
		"/var/run/lbost1ah.sock",
		"Healthcheck address. Example:'127.0.0.1:7000'")
	pflag.Duration("hc-timeout",
		2*time.Second,
		"Healthcheck request timeout")

	pflag.String("dummy-addr",
		"/var/run/lbost1ad.sock",
		"dummy address. Example:'/var/run/lbost1ad.sock'")
	pflag.Duration("dummy-timeout",
		2*time.Second,
		"dummy request timeout")
	pflag.String("route-addr",
		"/var/run/lbost1ar.sock",
		"route address. Example:'/var/run/lbost1ar.sock'")
	pflag.Duration("route-timeout",
		2*time.Second,
		"route request timeout")
	pflag.String("tunnel-addr",
		"/var/run/lbost1at.sock",
		"tunnel address. Example:'/var/run/lbost1at.sock'")
	pflag.Duration("tunnel-timeout",
		2*time.Second,
		"tunnel request timeout")
	pflag.String("ip-rule-addr",
		"/var/run/lbost1aipr.sock",
		"ip rule address. Example:'/var/run/lbost1aipr.sock'")
	pflag.Duration("ip-rule-timeout",
		2*time.Second,
		"ip rule request timeout")
	pflag.String("ipvs-addr",
		"/var/run/lbost1ai.sock",
		"ipvs address. Example:'/var/run/lbost1ai.sock'")
	pflag.Duration("ipvs-timeout",
		2*time.Second,
		"ipvs request timeout")

	pflag.String("consul-address",
		"127.0.0.1:18700",
		"consul address")
	pflag.String("consul-subscribe-path",
		"lbos/t1-cluster-1/",
		"consul subscribe path")
	pflag.String("consul-app-servers-path",
		"app-servers/",
		"consul app servers path")
	pflag.String("consul-manifest-name",
		"manifest",
		"consul service manifest")

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
	if viper.GetString("t1-id") == "bad1" {
		logging.Fatalf("t1 id must be set")
	}

	switch viper.GetString("id-type") {
	case "nanoid":
	case "id4":
	default:
		logging.Fatalf("unsuported id type: %v; supported types: nanoid|id4", viper.GetString("id-type"))
	}
}
