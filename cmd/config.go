package run

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thevan4/go-billet/logger"
)

const rootEntity = "root-entity"

// Default values
const (
	defaultConfigFilePath = "./nw-lb.yaml"
	defaultLogOutput      = "syslog"
	defaultLogLevel       = "trace"
	defaultLogFormat      = "default"
	defaultSystemLogTag   = ""

	defaultRestAPIIP   = "0.0.0.0"
	defaultRestAPIPort = "7000"

	defaultTechInterface = "" // must be set
	defaultFwmarkNumber  = "" // must be set
	// defaultPathToIfcfgTunnelFiles = "/etc/sysconfig/network-scripts/"
	// defaultSysctlConfigsPath      = "/etc/sysctl.d/"
	defaultDatabasePath           = "./database"
	defaultPathToIfcfgTunnelFiles = "/home/thevan/dev/goDev/prom/nw-pr-lb/network-scripts/"
	defaultSysctlConfigsPath      = "/home/thevan/dev/goDev/prom/nw-pr-lb/sysctl.d/"
	defaultMockMode               = true
	defaultHealtcheckTime         = 1 * time.Minute
	defaultMaxShutdownTimeName    = 20 * time.Second
)

// Config names
const (
	configFilePathName = "config-file-path"
	logOutputName      = "log-output"
	logLevelName       = "log-level"
	logFormatName      = "log-format"
	syslogTagName      = "syslog-tag"

	restAPIIPName   = "api-ip"
	restAPIPortName = "api-port"

	techInterfaceName          = "tech-interface"
	fwmarkNumberName           = "fwmark-number"
	pathToIfcfgTunnelFilesName = "path-to-ifcfg-tunnel-files"
	sysctlConfigsPathName      = "sysctl-configs-path"
	databasePathName           = "database-path"
	mockMode                   = "mock-mode"
	HealthcheckTimeName        = "validate-storage-config"
	maxShutdownTimeName        = "max-shutdown-time"
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

	pflag.String(restAPIIPName, defaultRestAPIIP, "Rest API ip")
	pflag.String(restAPIPortName, defaultRestAPIPort, "Rest API port")

	pflag.String(techInterfaceName, defaultTechInterface, "tech interface")
	pflag.String(fwmarkNumberName, defaultFwmarkNumber, "fwmark number")
	pflag.String(pathToIfcfgTunnelFilesName, defaultPathToIfcfgTunnelFiles, "path to ifcfg tunnel files")
	pflag.String(sysctlConfigsPathName, defaultSysctlConfigsPath, "sysctl config path")

	pflag.String(databasePathName, defaultDatabasePath, "Path to persistent database")
	pflag.Duration(HealthcheckTimeName, defaultHealtcheckTime, "Time interval for validate storage config")
	pflag.Duration(maxShutdownTimeName, defaultMaxShutdownTimeName, "Max time for graceful shutdown")

	pflag.Bool(mockMode, defaultMockMode, "Mock mode. No commands will be executed")

	pflag.Parse()
	viperConfig.BindPFlags(pflag.CommandLine)

	// work with config file
	viperConfig.SetConfigFile(viperConfig.GetString(configFilePathName))
	viperConfig.ReadInConfig()

	// init logs
	logging, err = logger.NewLogger(viperConfig.GetString(logOutputName),
		viperConfig.GetString(logLevelName),
		viperConfig.GetString(logFormatName),
		viperConfig.GetString(syslogTagName))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if viperConfig.GetString(techInterfaceName) == "" {
		logging.WithFields(logrus.Fields{
			"entity": rootEntity,
		}).Fatalf("tech interface must be set")
	}
	if viperConfig.GetString(fwmarkNumberName) == "" {
		logging.WithFields(logrus.Fields{
			"entity": rootEntity,
		}).Fatalf("fwmark number must be set")
	}
}
