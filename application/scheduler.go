package application

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const schedulerEntity = "scheduler"

// ValidateConfigScheduler ...
type ValidateConfigScheduler struct {
	storage       *portadapter.StorageEntity
	ipvsadmEntity *portadapter.IPVSADMEntity
	scheduler     *gocron.Scheduler
	locker        *domain.Locker
	signalChan    chan os.Signal
	logging       *logrus.Logger
}

// NewValidateConfigScheduler ...
func NewValidateConfigScheduler(storage *portadapter.StorageEntity,
	ipvsadmEntity *portadapter.IPVSADMEntity,
	locker *domain.Locker,
	signalChan chan os.Signal,
	logging *logrus.Logger) *ValidateConfigScheduler {
	return &ValidateConfigScheduler{
		storage:       storage,
		ipvsadmEntity: ipvsadmEntity,
		scheduler:     gocron.NewScheduler(time.UTC),
		locker:        locker,
		logging:       logging,
	}
}

// StartValidateConfigScheduler ...
func (validateConfigScheduler *ValidateConfigScheduler) StartValidateConfigScheduler(checkTime time.Duration) error {
	gocronCheckTime, err := transformTimeToUint64(checkTime)
	if err != nil {
		return fmt.Errorf("can't transform time to uint64: %v", err)
	}
	if gocronCheckTime != 1 {
		validateConfigScheduler.scheduler.Every(gocronCheckTime).Seconds().Do(validateConfigScheduler.validateConfig)
	} else {
		validateConfigScheduler.scheduler.Every(gocronCheckTime).Second().Do(validateConfigScheduler.validateConfig)
	}

	go validateConfigScheduler.startScheduler()
	return nil
}

func (validateConfigScheduler *ValidateConfigScheduler) startScheduler() {
	<-validateConfigScheduler.scheduler.Start()
}

func transformTimeToUint64(check time.Duration) (uint64, error) {
	timeSeconds := check.Seconds()
	s := fmt.Sprintf("%.0f", timeSeconds)
	return strconv.ParseUint(s, 10, 64)
}

func (validateConfigScheduler *ValidateConfigScheduler) validateConfig() {
	validateConfigScheduler.locker.Lock()
	defer validateConfigScheduler.locker.Unlock()
	err := validateConfigScheduler.ipvsadmEntity.ValidateHistoricalConfig(validateConfigScheduler.storage)
	if err != nil {
		validateConfigScheduler.logging.WithFields(logrus.Fields{
			"entity": schedulerEntity,
		}).Errorf("validate storage config by scheduller fail: %v", err)
		validateConfigScheduler.signalChan <- syscall.SIGINT
	}
}
