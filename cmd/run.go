package run

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/application"
	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thevan4/go-billet/executor"
)

var rootCmd = &cobra.Command{
	Use:   "nw-lb",
	Short: "network loadbalancer 😉",
	Run: func(cmd *cobra.Command, args []string) {
		// viperConfig, logging, uuidGenerator, uuidForRootProcess := prepareToStart()
		uuidGenerator := portadapter.NewUUIDGenerator()
		uuidForRootProcess := uuidGenerator.NewUUID().UUID.String()

		pathToKeepalivedDConfigConfigured := viperConfig.GetString(keepalivedConfigPathName) + "keepalived.d/services-configured/"
		pathToKeepalivedDConfigEnabled := viperConfig.GetString(keepalivedConfigPathName) + "keepalived.d/services-enabled/" // TODO: move that logic
		pathToKeepalivedConfig := viperConfig.GetString(keepalivedConfigPathName) + "keepalived.conf"

		// TODO: goreleaser - rpm packages create
		// TODO: rename map for remove names!
		// validate fields
		logging.WithFields(logrus.Fields{
			"entity":           rootEntity,
			"event uuid":       uuidForRootProcess,
			"Config file path": viperConfig.GetString(configFilePathName),
			"Log format":       viperConfig.GetString(logFormatName),
			"Log level":        viperConfig.GetString(logLevelName),
			"Log output":       viperConfig.GetString(logOutputName),
			"Syslog tag":       viperConfig.GetString(syslogTagName),

			"Rest API ip":   viperConfig.GetString(restAPIIPName),
			"Rest API port": viperConfig.GetString(restAPIPortName),

			"tech interface":                          viperConfig.GetString(techInterfaceName),
			"fwmark number":                           viperConfig.GetString(fwmarkNumberName),
			"path to ifcfg tunnel files":              viperConfig.GetString(pathToIfcfgTunnelFilesName),
			"sysctl config path":                      viperConfig.GetString(sysctlConfigsPathName),
			"keepalived folder path":                  viperConfig.GetString(keepalivedConfigPathName),
			"keepalived config path":                  pathToKeepalivedConfig,
			"keepalived daemon configured files path": pathToKeepalivedDConfigConfigured,
			"keepalived daemon enabled files path":    pathToKeepalivedDConfigEnabled,
			"mock mode":                               viperConfig.GetBool(mockMode),
		}).Info("")

		if isColdStart() && !viperConfig.GetBool(mockMode) {
			err := checkPrerequisites(pathToKeepalivedDConfigConfigured, pathToKeepalivedDConfigEnabled, pathToKeepalivedConfig)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"entity":     rootEntity,
					"event uuid": uuidForRootProcess,
				}).Fatalf("checkPrerequisites error: %v", err)
			}
		} // TODO: remove hardcode

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		nwConfig := domain.NewNetworkConfig()
		// nwConfig := domain.NewNetworkConfig(viperConfig.GetString(techInterfaceName),
		// 	viperConfig.GetString(fwmarkNumberName),
		// 	viperConfig.GetString(pathToIfcfgTunnelFilesName),
		// 	viperConfig.GetString(sysctlConfigsPathName),
		// 	pathToKeepalivedConfig,
		// 	pathToKeepalivedDConfigConfigured,
		// 	pathToKeepalivedDConfigEnabled)

		// tunnel maker start
		tunnelMaker := portadapter.NewTunnelFileMaker(viperConfig.GetString(pathToIfcfgTunnelFilesName),
			viperConfig.GetString(sysctlConfigsPathName),
			viperConfig.GetBool(mockMode),
			logging)
		// tunnel maker end
		// keepaliver maker start
		keepalivedCustomizer := portadapter.NewKeepalivedCustomizer(viperConfig.GetString(techInterfaceName),
			viperConfig.GetString(fwmarkNumberName),
			pathToKeepalivedConfig,
			pathToKeepalivedDConfigConfigured,
			pathToKeepalivedDConfigEnabled,
			viperConfig.GetBool(mockMode),
			logging)
		// keepaliver maker end
		facade := application.NewBalancerFacade(nwConfig, tunnelMaker, keepalivedCustomizer, uuidGenerator, logging)

		restAPI := application.NewRestAPIentity(viperConfig.GetString(restAPIIPName), viperConfig.GetString(restAPIPortName), facade)
		go restAPI.UpRestAPI()
		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
			// add some new here
		}).Info("Program stoped")
	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func isColdStart() bool { // TODO: write logic
	return true
}

func checkPrerequisites(pathToKeepalivedDConfigConfigured, pathToKeepalivedDConfigEnabled, pathToKeepalivedConfig string) error {

	if _, err := os.Stat(pathToKeepalivedDConfigConfigured); os.IsNotExist(err) {
		err = os.MkdirAll(pathToKeepalivedDConfigConfigured, 0744)
		if err != nil {
			return fmt.Errorf("can't make dir %v, got error %v", pathToKeepalivedDConfigConfigured, err)
		}
	}

	if _, err := os.Stat(pathToKeepalivedDConfigEnabled); os.IsNotExist(err) {
		err = os.MkdirAll(pathToKeepalivedDConfigEnabled, 0744)
		if err != nil {
			return fmt.Errorf("can't make dir %v, got error %v", pathToKeepalivedDConfigEnabled, err)
		}
	}

	if _, err := os.OpenFile(pathToKeepalivedConfig, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
		return fmt.Errorf("can't check/create file %v, got error %v", pathToKeepalivedConfig, err)
	}

	var err error
	dummyModprobeDPath := "/etc/modprobe.d/dummy.conf"         // TODO: remove hardcode
	expectDummyModprobContains := "options dummy numdummies=1" // TODO: remove hardcode
	if err = checkFileContains(dummyModprobeDPath, expectDummyModprobContains); err != nil {
		return fmt.Errorf("error when check dummy file: %v", err)
	}

	dummyModuleFilePath := "/etc/modules-load.d/dummy.conf" // TODO: remove hardcode
	expectDummyModuleFileContains := "dummy"                // TODO: remove hardcode
	if err := checkFileContains(dummyModuleFilePath, expectDummyModuleFileContains); err != nil {
		return fmt.Errorf("error when check dummy module file: %v", err)
	}

	tunnelModuleFilePath := "/etc/modules-load.d/tunnel.conf" // TODO: remove hardcode
	expectTunnelModuleFileContains := "tunnel4"               // TODO: remove hardcode
	if err := checkFileContains(tunnelModuleFilePath, expectTunnelModuleFileContains); err != nil {
		return fmt.Errorf("error when check tunnel module file: %v", err)
	}

	if err := checkKeepalivedIsRunning(true); err != nil {
		return fmt.Errorf("error when ceck keepalived service: %v", err)
	}

	return nil
}

func checkFileContains(filePath, expectedData string) error {
	dataBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("can't read file %v, got error %v", filePath, err)
	}
	if !strings.Contains(string(dataBytes), expectedData) {
		return fmt.Errorf("dummy file %v dosen't contains %v, only data: %v", filePath, expectedData, string(dataBytes))
	}
	return nil
}

func checkKeepalivedIsRunning(isFirstCheck bool) error {
	args := []string{"status", "keepalived"}
	_, _, exitCode, err := executor.Execute("/usr/bin/systemctl", "", args)
	if err != nil {
		return fmt.Errorf("when execute command %v, got error: %v", "/usr/bin/systemctl status keepalived", err)
	}
	if exitCode == 3 {
		if isFirstCheck {
			err := tryStartKeepalivedService()
			if err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("can't start keepalived service")
	}
	return nil
}

func tryStartKeepalivedService() error {
	args := []string{"start", "keepalived"}
	stdout, stderr, exitCode, err := executor.Execute("/usr/bin/systemctl", "", args)
	if err != nil {
		return fmt.Errorf("when execute command %v, got error: %v", "/usr/bin/systemctl start keepalived", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("when execute command %v, got exit code != 0: stdout: %v, stderr: %v, exitCode: %v",
			"/usr/bin/systemctl start keepalived",
			string(stdout),
			string(stderr),
			exitCode)
	}
	time.Sleep(5 * time.Second)
	err = checkKeepalivedIsRunning(false)
	if err != nil {
		return fmt.Errorf("can't start keepalived service: %v", err)
	}
	return nil
}

// TODO: include "keepalived.d/services-enabled/*.conf" check it too!

//TODO:
// long: bird peering autoset when cold cold start

// add and remove real servers
// if real 1 - exit error
// gen commands for real servers!
// response: report about created (ip real services created)
// register in controller

// long: firewall rules (iptables)

// NEXTTODEPLOY:
// контроллер: тоже что и сюда + у него информация о всех балансировщиках (статус, количество, регистрация, разрегистрация, загрузка(CPU+network(сессии в сек)), мониторинг балансировщиков)
// на основе зоны безы выбор подсеть (номер автономной сети) и на резать там VM
