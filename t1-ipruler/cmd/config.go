package run

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	logger "github.com/thevan4/logrus-wrapper"
)

// Default values
const (
	defaultConfigFilePath   = "./lbost1aipr.yaml"
	defaultLogOutput        = "stdout"
	defaultLogLevel         = "trace"
	defaultLogFormat        = "text"
	defaultSystemLogTag     = ""
	defaultLogEventLocation = true

	defaultIDType = "nanoid"

	defaultHCAddress = "/var/run/lbost1ao.sock"
	defaultHCTimeout = 2 * time.Second

	defaultRouteAddress = "/var/run/lbost1aipr.sock"
	defaultRouteTimeout = 2 * time.Second
)

// Config names
const (
	configFilePathName   = "config-file-path"
	logOutputName        = "log-output"
	logLevelName         = "log-level"
	logFormatName        = "log-format"
	syslogTagName        = "syslog-tag"
	logEventLocationName = "log-event-location"

	hcAddressName = "hc-address"
	hcTimeoutName = "hc-timeout"

	ipRuleAddressName = "ipRule-address"
	ipRuleTimeoutName = "ipRule-timeout"

	idTypeName = "id-type"
)

// // For builds with ldflags
var (
	version   = "unknown"
	buildTime = "unknown"
	// 	commit  = "TBD @ ldflags"
	// 	branch  = "TBD @ ldflags"
)

var (
	viperConfig *viper.Viper
	logging     *logrus.Logger
)

func init() {
	var err error
	viperConfig = viper.New()
	// work with env
	viperConfig.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viperConfig.AutomaticEnv()

	// work with flags
	pflag.StringP(configFilePathName, "c", defaultConfigFilePath, "Path to config file. Example value: './nw-lb.yaml'")
	pflag.String(logOutputName, defaultLogOutput, "Log output. Example values: 'stdout', 'syslog'")
	pflag.String(logLevelName, defaultLogLevel, "Log level. Example values: 'info', 'debug', 'trace'")
	pflag.String(logFormatName, defaultLogFormat, "Log format. Example values: 'text', 'json'")
	pflag.String(syslogTagName, defaultSystemLogTag, "Syslog tag. Example: 'trac-dgen'")
	pflag.Bool(logEventLocationName, defaultLogEventLocation, "Log event location (like python)")

	pflag.String(hcAddressName, defaultHCAddress, "hc address")
	pflag.Duration(hcTimeoutName, defaultHCTimeout, "hc timeout")

	pflag.String(ipRuleAddressName, defaultRouteAddress, "ipRule address")
	pflag.Duration(ipRuleTimeoutName, defaultRouteTimeout, "ipRule address")

	pflag.String(idTypeName, defaultIDType, "ID type(nanoid|uuid4)")

	pflag.Parse()
	if err := viperConfig.BindPFlags(pflag.CommandLine); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// work with config file
	viperConfig.SetConfigFile(viperConfig.GetString(configFilePathName))
	if err := viperConfig.ReadInConfig(); err != nil {
		fmt.Println(err)
	}

	// init logs
	newLogger := &logger.Logger{
		Output:           []string{viperConfig.GetString(logOutputName)},
		Level:            viperConfig.GetString(logLevelName),
		Formatter:        viperConfig.GetString(logFormatName),
		LogEventLocation: viperConfig.GetBool(logEventLocationName),
	}
	logging, err = logger.NewLogrusLogger(newLogger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch viperConfig.GetString(idTypeName) {
	case "nanoid":
	case "id4":
	default:
		logging.Fatalf("unsuported id type: %v; supported types: nanoid|id4", viperConfig.GetString(idTypeName))
	}
}