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

const rootEntity = "root-entity"

// Default values
const (
	defaultConfigFilePath   = "./nw-lb.yaml"
	defaultLogOutput        = "syslog"
	defaultLogLevel         = "trace"
	defaultLogFormat        = "text"
	defaultSystemLogTag     = ""
	defaultLogEventLocation = true

	defaultRestAPIIP   = "127.0.0.1"
	defaultRestAPIPort = "7000"

	defaultTechInterface       = "" // required
	defaultFwmarkNumber        = "" // required
	defaultSysctlConfigsPath   = "/etc/sysctl.d/"
	defaultDatabasePath        = "./database"
	defaultMockMode            = false
	defaultHealtcheckTime      = 1 * time.Minute
	defaultMaxShutdownTimeName = 20 * time.Second

	defaultMainSecret                = "" // required
	defaultMainSecretForRefresh      = "" // required
	defaultExpireTokenTime           = 12 * time.Hour
	defaultExpireTokenForRefreshTime = 96 * time.Hour
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

	restAPIIPName   = "api-ip"
	restAPIPortName = "api-port"

	techInterfaceName     = "tech-interface"
	fwmarkNumberName      = "fwmark-number"
	sysctlConfigsPathName = "sysctl-configs-path"
	databasePathName      = "database-path"
	mockMode              = "mock-mode"
	HealthcheckTimeName   = "validate-storage-config"
	maxShutdownTimeName   = "max-shutdown-time"

	mainSecretName                = "main-secret"
	mainSecretForRefreshName      = "main-secret-for-refresh"
	credentials                   = "credentials"
	expireTokenTimeName           = "expire token time"
	expireTokenForRefreshTimeName = "expire token for refresh time"
)

// // For builds with ldflags
// var (
// 	version = "TBD @ ldflags"
// 	commit  = "TBD @ ldflags"
// 	branch  = "TBD @ ldflags"
// )

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
	pflag.String(logFormatName, defaultLogFormat, "Log format. Example values: 'default', 'json'")
	pflag.String(syslogTagName, defaultSystemLogTag, "Syslog tag. Example: 'trac-dgen'")
	pflag.Bool(logEventLocationName, defaultLogEventLocation, "Log event location (like python)")

	pflag.String(restAPIIPName, defaultRestAPIIP, "Rest API ip")
	pflag.String(restAPIPortName, defaultRestAPIPort, "Rest API port")

	pflag.String(techInterfaceName, defaultTechInterface, "tech interface")
	pflag.String(fwmarkNumberName, defaultFwmarkNumber, "fwmark number")
	pflag.String(sysctlConfigsPathName, defaultSysctlConfigsPath, "sysctl config path")

	pflag.String(databasePathName, defaultDatabasePath, "Path to persistent database")
	pflag.Duration(HealthcheckTimeName, defaultHealtcheckTime, "Time interval for validate storage config")
	pflag.Duration(maxShutdownTimeName, defaultMaxShutdownTimeName, "Max time for graceful shutdown")

	pflag.Bool(mockMode, defaultMockMode, "Mock mode. No commands will be executed")

	pflag.String(mainSecretName, defaultMainSecret, "Main secret for JWT")
	pflag.String(mainSecretForRefreshName, defaultMainSecretForRefresh, "Refresh secret for JWT")
	pflag.StringToString(credentials, defaultCredentials, "User credentials")
	pflag.Duration(expireTokenTimeName, defaultExpireTokenTime, "Expire time for jwt token")
	pflag.Duration(expireTokenForRefreshTimeName, defaultExpireTokenForRefreshTime, "Expire time for refresh jwt token")

	pflag.Parse()
	viperConfig.BindPFlags(pflag.CommandLine)

	// work with config file
	viperConfig.SetConfigFile(viperConfig.GetString(configFilePathName))
	viperConfig.ReadInConfig()

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

	// required values are set
	if viperConfig.GetString(techInterfaceName) == "" {
		logging.Fatalf("tech interface must be set")
	}
	if viperConfig.GetString(fwmarkNumberName) == "" {
		logging.Fatalf("fwmark number must be set")
	}
	if viperConfig.GetString(mainSecretName) == "" {
		logging.Fatalf("secret for JWT number must be set")
	}
	if viperConfig.GetString(mainSecretForRefreshName) == "" {
		logging.Fatalf("refresh secret for JWT must be set")
	}
	if len(viperConfig.GetStringMapString(credentials)) == 0 {
		logging.Fatalf("credentials for JWT must be set")
	}
}
