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
	Use:   "nw-lb",
	Short: "network loadbalancer 😉",
	Run: func(cmd *cobra.Command, args []string) {
		// viperConfig, logging, uuidGenerator, uuidForRootProcess := prepareToStart()
		uuidGenerator := portadapter.NewUUIDGenerator()
		uuidForRootProcess := uuidGenerator.NewUUID().UUID.String()

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
			err := checkPrerequisites()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"entity":     rootEntity,
					"event uuid": uuidForRootProcess,
				}).Fatalf("checkPrerequisites error: %v", err)
			}
		} // TODO: remove hardcode

		locker := &domain.Locker{}
		gracefullShutdown := &domain.GracefullShutdown{}

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
		cacheDB, persistentDB, err := storageAndCacheInit(viperConfig.GetString(databasePathName), logging)
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

		//  healthchecks start
		hc := usecase.NewHeathcheckEntity(cacheDB,
			persistentDB,
			ipvsadmConfigurator,
			viperConfig.GetString(techInterfaceName),
			locker,
			gracefullShutdown,
			viperConfig.GetBool(mockMode),
			logging)
		if err = hc.StartHealthchecksForCurrentServices(); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("Fail to load storage data to services info for healthcheck")
		}
		go hc.StartGracefullShutdownControlForHealthchecks()
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
			gracefullShutdown,
			uuidGenerator,
			logging)

		restAPI := application.NewRestAPIentity(viperConfig.GetString(restAPIIPName), viperConfig.GetString(restAPIPortName), facade)
		go restAPI.UpRestAPI()
		go restAPI.GracefulShutdownRestAPI(gracefulShutdownCommandForRestAPI, restAPIisDone)

		<-signalChan // shutdown signal
		if err := ipvsadmConfigurator.Flush(); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("IPVSADM can't flush data at stop: %v", err)
		}
		gracefulShutdownCommandForRestAPI <- struct{}{}
		gracefulShutdownUsecases(gracefullShutdown, viperConfig.GetDuration(maxShutdownTimeName), logging)
		<-restAPIisDone

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("Program stoped")
	},
}

func gracefulShutdownUsecases(gracefullShutdown *domain.GracefullShutdown, maxWaitTimeForJobsIsDone time.Duration, logging *logrus.Logger) {
	gracefullShutdown.Lock()
	gracefullShutdown.ShutdownNow = true
	gracefullShutdown.Unlock()

	ticker := time.NewTicker(time.Duration(100 * time.Millisecond)) // hardcode
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTimeForJobsIsDone)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			gracefullShutdown.Lock()
			if gracefullShutdown.UsecasesJobs <= 0 {
				logging.WithFields(logrus.Fields{
					"entity": rootEntity,
				}).Info("All jobs is done")
				defer gracefullShutdown.Unlock()
				return
			}
			gracefullShutdown.Unlock()
			continue
		case <-ctx.Done():
			gracefullShutdown.Lock()
			logging.WithFields(logrus.Fields{
				"entity": rootEntity,
			}).Warnf("%v jobs is fail when program stop", gracefullShutdown.UsecasesJobs)
			defer gracefullShutdown.Unlock()
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

func isColdStart() bool { // TODO: write logic
	return true
}

func checkPrerequisites() error {
	var err error
	dummyModprobeDPath := "/etc/modprobe.d/dummy.conf" // TODO: remove hardcode
	expectDummyModprobContains := "numdummies=1"       // TODO: remove hardcode
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

func storageAndCacheInit(databasePath string, logging *logrus.Logger) (*portadapter.StorageEntity, *portadapter.StorageEntity, error) {
	cacheDB, err := portadapter.NewStorageEntity(true, "", logging)
	if err != nil {
		return nil, nil, fmt.Errorf("init NewStorageEntity for cache error: %v", err)
	}

	persistentDB, err := portadapter.NewStorageEntity(false, databasePath, logging)
	if err != nil {
		return nil, nil, fmt.Errorf("init NewStorageEntity for persistent storage error: %v", err)
	}

	err = cacheDB.LoadCacheFromStorage(persistentDB)
	if err != nil {
		return nil, nil, fmt.Errorf("cant load cache from storage: %v", err)
	}

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

// TODO: long: bird peering autoset when cold cold start
