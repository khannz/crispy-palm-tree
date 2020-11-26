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
	defaultConfigFilePath   = "./lbost1ao.yaml"
	defaultLogOutput        = "stdout"
	defaultLogLevel         = "trace"
	defaultLogFormat        = "text"
	defaultSystemLogTag     = ""
	defaultLogEventLocation = true

	defaultT1OrchID      = "" // required
	defaultHlckInterface = "" // required

	defaultIDType = "nanoid"

	// FIXME: 100500 addresses
	defaultHCAddress = "/var/run/lbost1ah.sock"
	defaultHCTimeout = 2 * time.Second

	defaultTunSockAddr = "/var/run/lbost1at.sock"
	defaultTunTimeout  = 2 * time.Second
)

var defaultCredentials = map[string]string{}

// Config names
const (
	configFilePathName   = "config-file-path"
	logOutputName        = "log-output"
	logLevelName         = "log-level"
	logFormatName        = "log-format"
	syslogTagName        = "syslog-tag"
	logEventLocationName = "log-event-location"

	t1OrchIDName      = "t1-id"
	hlckInterfaceName = "hlck-interface"

	maxShutdownTimeName = "max-shutdown-time"

	idTypeName = "id-type"

	hcAddressName = "hc-address"
	hcTimeoutName = "hc-timeout"

	tunSockAddrName = "tun-sock-addr"
	tunTimeoutName  = "tun-timeout"
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

	pflag.String(t1OrchIDName, defaultT1OrchID, "t1 orch id")

	pflag.String(defaultHlckInterface, hlckInterfaceName, "healthcheck interface")

	pflag.String(idTypeName, defaultIDType, "ID type(nanoid|uuid4)")

	pflag.String(hcAddressName, defaultHCAddress, "Healthcheck address. Example:'127.0.0.1:7000'")
	pflag.Duration(hcTimeoutName, defaultHCTimeout, "Healthcheck request timeout")

	pflag.String(tunSockAddrName, defaultTunSockAddr, "tunnel address. Example:'/var/run/lbost1at.sock'")
	pflag.Duration(tunTimeoutName, defaultTunTimeout, "tunnel request timeout")

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

	// required values are set // FIXME: set
	// if viperConfig.GetString(techInterfaceName) == "" {
	// 	logging.Fatalf("tech interface must be set")
	// }
	// if viperConfig.GetString(t1OrchIDName) == "" {
	// 	logging.Fatalf("t1 orch id must be set")
	// }

	switch viperConfig.GetString(idTypeName) {
	case "nanoid":
	case "id4":
	default:
		logging.Fatalf("unsuported id type: %v; supported types: nanoid|id4", viperConfig.GetString(idTypeName))
	}
}