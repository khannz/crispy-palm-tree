package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestGetAllServices ...
func TestGetAllServices(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	gracefulShutdown := &domain.GracefulShutdown{
		ShutdownNow:  true,
		UsecasesJobs: 0,
	}
	newLogger := &logger.Logger{
		Output:           []string{"stdout"},
		Level:            "trace",
		Formatter:        "text",
		LogEventLocation: true,
	}
	logging, errCreateLogging := logger.NewLogrusLogger(newLogger)
	assert.Nil(errCreateLogging)
	mockHeathcheckEntity := &MockHCWorker{}

	getAllServicesErr := NewGetAllServices(locker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	_, errNotNilOne := getAllServicesErr.GetAllServices("")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	_, errNotNilTwo := getAllServicesErr.GetAllServices("")
	assert.Nil(errNotNilTwo)

}
