package run

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/khannz/crispy-palm-tree/application"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "lbost1ac 😉",
	Run: func(cmd *cobra.Command, args []string) {
		idGenerator := chooseIDGenerator(viperConfig.GetString(idTypeName))
		idForRootProcess := idGenerator.NewID()

		// validate fields
		logging.WithFields(logrus.Fields{
			"version":          version,
			"build time":       buildTime,
			"event id":         idForRootProcess,
			"config file path": viperConfig.GetString(configFilePathName),
			"log format":       viperConfig.GetString(logFormatName),
			"log level":        viperConfig.GetString(logLevelName),
			"log output":       viperConfig.GetString(logOutputName),
			"syslog tag":       viperConfig.GetString(syslogTagName),

			"rest API ip":   viperConfig.GetString(restAPIIPName),
			"rest API port": viperConfig.GetString(restAPIPortName),

			"tech interface":                viperConfig.GetString(techInterfaceName),
			"fwmark number":                 viperConfig.GetString(fwmarkNumberName),
			"sysctl config path":            viperConfig.GetString(sysctlConfigsPathName),
			"database path":                 viperConfig.GetString(databasePathName),
			"mock mode":                     viperConfig.GetBool(mockMode),
			"time interval for healthcheck": viperConfig.GetDuration(HealthcheckTimeName),
			"max shutdown time":             viperConfig.GetDuration(maxShutdownTimeName),

			"expire token time":         viperConfig.GetDuration(expireTokenTimeName),
			"expire refresh token time": viperConfig.GetDuration(expireTokenForRefreshTimeName),
			"number of users":           len(viperConfig.GetStringMapString(credentials)),
			"id type":                   viperConfig.GetString(idTypeName),
			"hc address":                viperConfig.GetString(hcAddressName),
			"hc timeout":                viperConfig.GetDuration(hcTimeoutName),
		}).Info("")

		// FIXME: uncomment
		// if isColdStart() && !viperConfig.GetBool(mockMode) {
		// 	err := checkPrerequisites(idForRootProcess, logging)
		// 	if err != nil {
		// 		logging.WithFields(logrus.Fields{
		// 			"event id": idForRootProcess,
		// 		}).Fatalf("check prerequisites error: %v", err)
		// 	}
		// }

		locker := &domain.Locker{}
		gracefulShutdown := &domain.GracefulShutdown{}

		gracefulShutdownCommandForRestAPI := make(chan struct{}, 1)
		restAPIisDone := make(chan struct{}, 1)

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// tunnel maker start
		tunnelMaker := portadapter.NewTunnelFileMaker(viperConfig.GetString(sysctlConfigsPathName), // TODO: refactor. tunnels, routes and sysctl config maker
			viperConfig.GetBool(mockMode),
			logging)
		// tunnel maker end

		// db and caches init
		cacheDB, persistentDB, err := storageAndCacheInit(viperConfig.GetString(databasePathName), idForRootProcess, logging)
		if err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("storage and cache init error: %v", err)
		}
		defer cacheDB.Db.Close()
		defer persistentDB.Db.Close()

		// CommandGenerator start
		commandGenerator := portadapter.NewCommandGenerator()
		// CommandGenerator end

		//  healthchecks start
		hc := portadapter.NewHeathcheckEntity(viperConfig.GetString(hcAddressName), viperConfig.GetDuration(hcTimeoutName), logging)
		if err := hc.ConnectToHealtchecks(); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Fatalf("can't connect to healtchecks: %v", err)
		}
		// healthchecks end

		// init config end

		facade := application.NewBalancerFacade(locker,
			cacheDB,
			persistentDB,
			tunnelMaker,
			hc,
			commandGenerator,
			gracefulShutdown,
			idGenerator,
			logging)

		if err := facade.InitializeRuntimeSettings(idForRootProcess); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Errorf("initialize runtime settings fail: %v", err)
			if err := facade.DisableRuntimeSettings(viperConfig.GetBool(mockMode), idForRootProcess); err != nil {
				logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Errorf("disable runtime settings fail: %v", err)
			}
			os.Exit(1)
		}
		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("initialize runtime settings successful")

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Debug("healthchecks for current services started")

		// up rest api
		authorization := application.NewAuthorization(viperConfig.GetString(mainSecretName),
			viperConfig.GetString(mainSecretForRefreshName),
			viperConfig.GetStringMapString(credentials),
			viperConfig.GetDuration(expireTokenTimeName),
			viperConfig.GetDuration(expireTokenForRefreshTimeName))

		restAPI := application.NewRestAPIentity(viperConfig.GetString(restAPIIPName), viperConfig.GetString(restAPIPortName), authorization, facade, logging)
		go restAPI.UpRestAPI()
		go restAPI.GracefulShutdownRestAPI(gracefulShutdownCommandForRestAPI, restAPIisDone)

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("program running")

		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("got shutdown signal")

		gracefulShutdownCommandForRestAPI <- struct{}{}
		gracefulShutdownUsecases(gracefulShutdown, viperConfig.GetDuration(maxShutdownTimeName), logging)
		<-restAPIisDone
		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("rest API is Done")

		// gracefulShutdown.ShutdownNow = false // TODO: so dirty trick
		if err := facade.DisableRuntimeSettings(viperConfig.GetBool(mockMode), idForRootProcess); err != nil {
			logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Warnf("disable runtime settings errors: %v", err)
		}

		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("runtime settings disabled")
		hc.DisconnectFromHealtchecks()
		logging.WithFields(logrus.Fields{"event id": idForRootProcess}).Info("program stopped")
	},
}

func gracefulShutdownUsecases(gracefulShutdown *domain.GracefulShutdown, maxWaitTimeForJobsIsDone time.Duration, logging *logrus.Logger) {
	gracefulShutdown.Lock()
	gracefulShutdown.ShutdownNow = true
	gracefulShutdown.Unlock()

	ticker := time.NewTicker(time.Duration(100 * time.Millisecond)) // hardcode
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTimeForJobsIsDone)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			gracefulShutdown.Lock()
			if gracefulShutdown.UsecasesJobs <= 0 {
				logging.Info("All jobs is done")
				defer gracefulShutdown.Unlock()
				return
			}
			gracefulShutdown.Unlock()
			continue
		case <-ctx.Done():
			gracefulShutdown.Lock()
			logging.Warnf("%v jobs is fail when program stop", gracefulShutdown.UsecasesJobs)
			defer gracefulShutdown.Unlock()
			return
		}
	}
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func isColdStart() bool { // TODO: write logic?
	return true
}

func checkPrerequisites(id string, logging *logrus.Logger) error {
	logging.WithFields(logrus.Fields{"event id": id}).Info("start check prerequisites")
	var err error
	dummyModprobeDPath := "/etc/modprobe.d/dummy.conf" // TODO: remove hardcode?
	expectDummyModprobContains := "numdummies=1"       // TODO: remove hardcode?
	if err = checkFileContains(dummyModprobeDPath, expectDummyModprobContains); err != nil {
		return fmt.Errorf("error when check dummy file: %v", err)
	}
	logging.WithFields(logrus.Fields{"event id": id}).Debugf("check prerequisites in %v successful", dummyModprobeDPath)

	dummyModuleFilePath := "/etc/modules-load.d/lbos.conf" // TODO: remove hardcode?
	expectDummyModuleFileContains := "dummy"               // TODO: remove hardcode?
	if err := checkFileContains(dummyModuleFilePath, expectDummyModuleFileContains); err != nil {
		return fmt.Errorf("error when check dummy module file: %v", err)
	}
	logging.WithFields(logrus.Fields{"event id": id}).Debugf("check prerequisites in %v successful", dummyModuleFilePath)

	tunnelModuleFilePath := "/etc/modules-load.d/lbos.conf" // TODO: remove hardcode?
	expectTunnelModuleFileContains := "tunnel4"             // TODO: remove hardcode?
	if err := checkFileContains(tunnelModuleFilePath, expectTunnelModuleFileContains); err != nil {
		return fmt.Errorf("error when check tunnel module file: %v", err)
	}
	logging.WithFields(logrus.Fields{"event id": id}).Debugf("check prerequisites in %v successful", tunnelModuleFilePath)

	logging.WithFields(logrus.Fields{"event id": id}).Info("check all prerequisites successful")
	// check all:
	// ip_vs
	// dummy
	// ipip
	// tunnel4
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

func storageAndCacheInit(databasePath, id string, logging *logrus.Logger) (*portadapter.StorageEntity, *portadapter.StorageEntity, error) {
	cacheDB, err := portadapter.NewStorageEntity(true, "", logging)
	if err != nil {
		return nil, nil, fmt.Errorf("init NewStorageEntity for cache error: %v", err)
	}
	logging.WithFields(logrus.Fields{"event id": id}).Debug("init cacheDB successful")

	persistentDB, err := portadapter.NewStorageEntity(false, databasePath, logging)
	if err != nil {
		return nil, nil, fmt.Errorf("init NewStorageEntity for persistent storage error: %v", err)
	}
	logging.WithFields(logrus.Fields{"event id": id}).Debug("init persistentDB successful")

	err = cacheDB.LoadCacheFromStorage(persistentDB)
	if err != nil {
		return nil, nil, fmt.Errorf("cant load cache from storage: %v", err)
	}
	logging.WithFields(logrus.Fields{"event id": id}).Debug("load cache from persistent storage successful")

	return cacheDB, persistentDB, nil
}

func chooseIDGenerator(idType string) domain.IDgenerator {
	switch viperConfig.GetString(idTypeName) {
	case "nanoid":
		return portadapter.NewIDGenerator()
	case "uuid4":
		return portadapter.NewUUIIDGenerator()
	default:
		return portadapter.NewIDGenerator()
	}
}
