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
	"github.com/khannz/crispy-palm-tree/usecase"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lb",
	Short: "load balancer tier 1 😉",
	Run: func(cmd *cobra.Command, args []string) {
		// viperConfig, logging, uuidGenerator, uuidForRootProcess := prepareToStart()
		uuidGenerator := portadapter.NewUUIDGenerator()
		uuidForRootProcess := uuidGenerator.NewUUID().UUID.String()

		// validate fields
		logging.WithFields(logrus.Fields{
			"entity":           rootEntity,
			"event uuid":       uuidForRootProcess,
			"config file path": viperConfig.GetString(configFilePathName),
			"log format":       viperConfig.GetString(logFormatName),
			"log level":        viperConfig.GetString(logLevelName),
			"log output":       viperConfig.GetString(logOutputName),
			"syslog tag":       viperConfig.GetString(syslogTagName),

			"rest API ip":   viperConfig.GetString(restAPIIPName),
			"rest API port": viperConfig.GetString(restAPIPortName),

			"tech interface":                viperConfig.GetString(techInterfaceName),
			"fwmark number":                 viperConfig.GetString(fwmarkNumberName),
			"path to ifcfg tunnel files":    viperConfig.GetString(pathToIfcfgTunnelFilesName),
			"sysctl config path":            viperConfig.GetString(sysctlConfigsPathName),
			"database path":                 viperConfig.GetString(databasePathName),
			"mock mode":                     viperConfig.GetBool(mockMode),
			"time interval for healthcheck": viperConfig.GetDuration(HealthcheckTimeName),
			"max shutdown time":             viperConfig.GetDuration(maxShutdownTimeName),
		}).Info("")

		if isColdStart() && !viperConfig.GetBool(mockMode) {
			err := checkPrerequisites(uuidForRootProcess, logging)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"entity":     rootEntity,
					"event uuid": uuidForRootProcess,
				}).Fatalf("check prerequisites error: %v", err)
			}
		}

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
		tunnelMaker := portadapter.NewTunnelFileMaker(viperConfig.GetString(pathToIfcfgTunnelFilesName),
			viperConfig.GetString(sysctlConfigsPathName),
			viperConfig.GetBool(mockMode),
			logging)
		// tunnel maker end

		// db and caches init
		cacheDB, persistentDB, err := storageAndCacheInit(viperConfig.GetString(databasePathName), uuidForRootProcess, logging)
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("storage and cache init error: %v", err)
		}
		defer cacheDB.Db.Close()
		defer persistentDB.Db.Close()

		// ipvsadmConfigurator start
		ipvsadmConfigurator, err := portadapter.NewIPVSADMEntity()
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("can't create IPVSADM entity: %v", err)
		}
		if err := ipvsadmConfigurator.Flush(); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("IPVSADM can't flush data at start: %v", err)
		}
		// ipvsadmConfigurator end

		// CommandGenerator start
		commandGenerator := portadapter.NewCommandGenerator()
		// CommandGenerator end

		//  healthchecks start
		hc := usecase.NewHeathcheckEntity(cacheDB,
			persistentDB,
			ipvsadmConfigurator,
			viperConfig.GetString(techInterfaceName),
			locker,
			gracefulShutdown,
			viperConfig.GetBool(mockMode),
			logging)
		if err = hc.StartHealthchecksForCurrentServices(); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("Fail to load storage data to services info for healthcheck: %v", err)
		}
		go hc.StartGracefulShutdownControlForHealthchecks() // TODO: graceful shutdown for healthchecks is overhead. Remove that?
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Debug("healthchecks for current services started")
		// healthchecks end

		// init config start
		if err := initConfigFromStorage(ipvsadmConfigurator, cacheDB, rootEntity); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("init config from storage fail: %v", err)
		}
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Debug("init storage config successful")

		// init config end

		facade := application.NewBalancerFacade(locker,
			ipvsadmConfigurator,
			cacheDB,
			persistentDB,
			tunnelMaker,
			hc,
			commandGenerator,
			gracefulShutdown,
			uuidGenerator,
			logging)

		restAPI := application.NewRestAPIentity(viperConfig.GetString(restAPIIPName), viperConfig.GetString(restAPIPortName), facade)
		go restAPI.UpRestAPI()
		go restAPI.GracefulShutdownRestAPI(gracefulShutdownCommandForRestAPI, restAPIisDone)

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("program running")

		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("got shutdown signal")

		if err := ipvsadmConfigurator.Flush(); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("IPVSADM can't flush data at stop: %v", err)
		}
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("IPVSADM data has flushed")

		gracefulShutdownCommandForRestAPI <- struct{}{}
		gracefulShutdownUsecases(gracefulShutdown, viperConfig.GetDuration(maxShutdownTimeName), logging)
		<-restAPIisDone
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("rest API is Done")

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("program stopped")
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
				logging.WithFields(logrus.Fields{
					"entity": rootEntity,
				}).Info("All jobs is done")
				defer gracefulShutdown.Unlock()
				return
			}
			gracefulShutdown.Unlock()
			continue
		case <-ctx.Done():
			gracefulShutdown.Lock()
			logging.WithFields(logrus.Fields{
				"entity": rootEntity,
			}).Warnf("%v jobs is fail when program stop", gracefulShutdown.UsecasesJobs)
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

func checkPrerequisites(uuid string, logging *logrus.Logger) error {
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Info("start check prerequisites")
	var err error
	dummyModprobeDPath := "/etc/modprobe.d/dummy.conf" // TODO: remove hardcode?
	expectDummyModprobContains := "numdummies=1"       // TODO: remove hardcode?
	if err = checkFileContains(dummyModprobeDPath, expectDummyModprobContains); err != nil {
		return fmt.Errorf("error when check dummy file: %v", err)
	}
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Debugf("check prerequisites in %v successful", dummyModprobeDPath)

	dummyModuleFilePath := "/etc/modules-load.d/dummy.conf" // TODO: remove hardcode?
	expectDummyModuleFileContains := "dummy"                // TODO: remove hardcode?
	if err := checkFileContains(dummyModuleFilePath, expectDummyModuleFileContains); err != nil {
		return fmt.Errorf("error when check dummy module file: %v", err)
	}
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Debugf("check prerequisites in %v successful", dummyModuleFilePath)

	tunnelModuleFilePath := "/etc/modules-load.d/tunnel.conf" // TODO: remove hardcode?
	expectTunnelModuleFileContains := "tunnel4"               // TODO: remove hardcode?
	if err := checkFileContains(tunnelModuleFilePath, expectTunnelModuleFileContains); err != nil {
		return fmt.Errorf("error when check tunnel module file: %v", err)
	}
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Debugf("check prerequisites in %v successful", tunnelModuleFilePath)

	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Info("check all prerequisites successful")

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

func storageAndCacheInit(databasePath, uuid string, logging *logrus.Logger) (*portadapter.StorageEntity, *portadapter.StorageEntity, error) {
	cacheDB, err := portadapter.NewStorageEntity(true, "", logging)
	if err != nil {
		return nil, nil, fmt.Errorf("init NewStorageEntity for cache error: %v", err)
	}
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Debug("init cacheDB successful")

	persistentDB, err := portadapter.NewStorageEntity(false, databasePath, logging)
	if err != nil {
		return nil, nil, fmt.Errorf("init NewStorageEntity for persistent storage error: %v", err)
	}
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Debug("init cacheDB successful")

	err = cacheDB.LoadCacheFromStorage(persistentDB)
	if err != nil {
		return nil, nil, fmt.Errorf("cant load cache from storage: %v", err)
	}
	logging.WithFields(logrus.Fields{
		"entity":     rootEntity,
		"event uuid": uuid,
	}).Debug("load cache from persistent storage successful")

	return cacheDB, persistentDB, nil
}

func initConfigFromStorage(ipvsadmConfigurator *portadapter.IPVSADMEntity,
	storage *portadapter.StorageEntity,
	eventUUID string) error {
	configsFromStorage, err := storage.LoadAllStorageDataToDomainModel()
	if err != nil {
		return fmt.Errorf("fail to load  storage config at start")
	}
	for _, configFromStorage := range configsFromStorage {
		if err := ipvsadmConfigurator.CreateService(configFromStorage, eventUUID); err != nil {
			return fmt.Errorf("can't create service for %v config from storage: %v", configFromStorage, err)
		}
	}
	return nil
}

// TODO: long: bird peering autoset when cold cold start?
