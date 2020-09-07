package usecase

import (
	"testing"
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestValidateRemoveApplicationServers ...
func TestValidateRemoveApplicationServers(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, applicattionServersForRemove, appServer := createApplicationServersForTests()

	errNotNilOne := validateRemoveApplicationServers(currentApplicattionServers, applicattionServersForRemove)
	assert.NotNil(errNotNilOne)

	errNilOne := validateRemoveApplicationServers(currentApplicattionServers, []*domain.ApplicationServer{appServer})
	assert.Nil(errNilOne)

	errNotNilTwo := validateRemoveApplicationServers(currentApplicattionServers, []*domain.ApplicationServer{applicattionServersForRemove[1]})
	assert.NotNil(errNotNilTwo)

	newApplicattionServersForRemove := []*domain.ApplicationServer{applicattionServersForRemove[1], appServer}
	errNotNilThree := validateRemoveApplicationServers(currentApplicattionServers, newApplicattionServersForRemove)
	assert.NotNil(errNotNilThree)

	newApplicattionServersForRemove = append(newApplicattionServersForRemove, applicattionServersForRemove[0])
	errNotNilFour := validateRemoveApplicationServers(currentApplicattionServers, newApplicattionServersForRemove)
	assert.NotNil(errNotNilFour)
}

// TestFormNewApplicationServersSlice ...
func TestFormNewApplicationServersSlice(t *testing.T) {
	currentApplicattionServers, _, appServer := createApplicationServersForTests()

	reNewApplicattionServers := formNewApplicationServersSlice(currentApplicattionServers, []*domain.ApplicationServer{appServer})
	for _, reNewApplicattionServer := range reNewApplicattionServers {
		if reNewApplicattionServer.ServerIP == appServer.ServerIP && reNewApplicattionServer.ServerPort == appServer.ServerPort {
			t.Errorf("application server was not excluded from slice: %v:%v", reNewApplicattionServer.ServerIP, reNewApplicattionServer.ServerPort)
		}
	}
}

// TestForAddApplicationServersFormUpdateServiceInfo
func TestForAddApplicationServersFormUpdateServiceInfo(t *testing.T) {
	currentApplicattionServers, newApplicattionServers, _ := createApplicationServersForTests()
	serviceInfoOne, serviceInfoTwo, _ := createServicesInfoForTests(currentApplicattionServers, newApplicattionServers)
	_, nilErrOne := forAddApplicationServersFormUpdateServiceInfo(serviceInfoOne, serviceInfoTwo, "")
	assert.Nil(t, nilErrOne)
}

// TestForRemoveApplicationServersFormUpdateServiceInfo ...
func TestForRemoveApplicationServersFormUpdateServiceInfo(t *testing.T) {
	currentApplicattionServers, tmpApplicattionServers, appServerForRemove := createApplicationServersForTests()
	serviceInfoOne, _, serviceInfoThree := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)
	serviceInfoThree.ApplicationServers = []*domain.ApplicationServer{appServerForRemove}
	forRemoveApplicationServersFormUpdateServiceInfo(serviceInfoOne, serviceInfoThree, "")
}

// TestCheckRoutingTypeForApplicationServersValid ...
func TestCheckRoutingTypeForApplicationServersValid(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, appServerForRemove := createApplicationServersForTests()
	serviceInfoOne, _, serviceInfoThree := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)
	serviceInfoThree.ApplicationServers = []*domain.ApplicationServer{appServerForRemove}
	// newServiceInfo *domain.ServiceInfo, allCurrentServices []*domain.ServiceInfo
	nilErr := checkRoutingTypeForApplicationServersValid(serviceInfoThree, []*domain.ServiceInfo{serviceInfoOne})
	assert.Nil(nilErr)

	serviceInfoThree.RoutingType = "masquerading"
	notNilErr := checkRoutingTypeForApplicationServersValid(serviceInfoThree, []*domain.ServiceInfo{serviceInfoOne})
	assert.NotNil(notNilErr)
}

func TestCheckServiceIPAndPortUnique(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	notNilErrOne := checkServiceIPAndPortUnique("111.111.111.111", "111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(notNilErrOne)

	notNilErrTwo := checkServiceIPAndPortUnique("1.1.1.1", "1111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(notNilErrTwo)

	nilErr := checkServiceIPAndPortUnique("9.1.1.1", "9111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.Nil(nilErr)
}

// TestDecreaseJobs ...
func TestDecreaseJobs(t *testing.T) {
	gs := &domain.GracefulShutdown{
		ShutdownNow:  false,
		UsecasesJobs: 1,
	}
	decreaseJobs(gs)
}

// TestFormTunnelsFilesInfo ...
func TestFormTunnelsFilesInfo(t *testing.T) {
	assert := assert.New(t)
	newLogger := &logger.Logger{
		Output:           []string{"stdout"},
		Level:            "trace",
		Formatter:        "text",
		LogEventLocation: true,
	}
	logging, errCreateLogging := logger.NewLogrusLogger(newLogger)
	assert.Nil(errCreateLogging)

	cacheDB, errCreateCacheDB := portadapter.NewStorageEntity(true, "", logging)
	assert.Nil(errCreateCacheDB)
	testUUID := "test uuid 1"

	currentApplicattionServers, tmpApplicattionServers, appServer := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	errPutToDb := cacheDB.NewServiceDataToStorage(currentServiceInfoOne, testUUID)
	assert.Nil(errPutToDb)

	FormTunnelsFilesInfo([]*domain.ApplicationServer{appServer}, cacheDB)
}

// TestCheckApplicationServersIPAndPortUnique ...
func TestCheckApplicationServersIPAndPortUnique(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, appServer := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	errNotNilOne := checkApplicationServersIPAndPortUnique([]*domain.ApplicationServer{appServer}, []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(errNotNilOne)

	appServer.ServerIP = "111.111.111.111"
	appServer.ServerPort = "111"
	errNotNilTwo := checkApplicationServersIPAndPortUnique([]*domain.ApplicationServer{appServer}, []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(errNotNilTwo)

	appServer.ServerIP = "91.1.1.1"
	appServer.ServerPort = "911"
	errNil := checkApplicationServersIPAndPortUnique([]*domain.ApplicationServer{appServer}, []*domain.ServiceInfo{currentServiceInfoOne})
	assert.Nil(errNil)
}

// TestIsServiceExist ...
func TestIsServiceExist(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	notExist := isServiceExist("91.1.1.1", "9111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.False(notExist)

	exist := isServiceExist("111.111.111.111", "111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.True(exist)
}

// TestCheckApplicationServersExistInService ...
func TestCheckApplicationServersExistInService(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, appServer := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	errNil := checkApplicationServersExistInService([]*domain.ApplicationServer{appServer}, currentServiceInfoOne)
	assert.Nil(errNil)

	appServer.ServerIP = "91.1.1.1"
	appServer.ServerPort = "911"
	errNotNil := checkApplicationServersExistInService([]*domain.ApplicationServer{appServer}, currentServiceInfoOne)
	assert.NotNil(errNotNil)
}

func createServicesInfoForTests(applicattionServersOne []*domain.ApplicationServer, applicattionServersTwo []*domain.ApplicationServer) (*domain.ServiceInfo, *domain.ServiceInfo, *domain.ServiceInfo) {
	serviceHealthcheckOne, serviceHealthcheckTwo, _ := createServicesHealthchecksForTests()
	serviceInfoOne := &domain.ServiceInfo{
		ServiceIP:          "111.111.111.111",
		ServicePort:        "111",
		ApplicationServers: applicattionServersOne,
		Healthcheck:        serviceHealthcheckOne,
		// ExtraInfo          []string             `json:"extraInfo"`
		IsUp:        true,
		BalanceType: "rr",
		RoutingType: "tunneling",
	}
	serviceInfoTwo := &domain.ServiceInfo{
		ServiceIP:          "222.222.222.222",
		ServicePort:        "222",
		ApplicationServers: applicattionServersOne,
		Healthcheck:        serviceHealthcheckTwo,
		// ExtraInfo          []string             `json:"extraInfo"`
		IsUp:        true,
		BalanceType: "rr",
		RoutingType: "tunneling",
	}
	serviceInfoThree := &domain.ServiceInfo{
		ServiceIP:   "111.111.111.111",
		ServicePort: "111",
		BalanceType: "rr",
		RoutingType: "tunneling",
	}

	return serviceInfoOne, serviceInfoTwo, serviceInfoThree
}

func createServicesHealthchecksForTests() (domain.ServiceHealthcheck, domain.ServiceHealthcheck, domain.ServiceHealthcheck) {
	serviceHealthcheckOne := domain.ServiceHealthcheck{
		StopChecks:           make(chan struct{}),
		PercentOfAlivedForUp: 50,
		Type:                 "http",
		Timeout:              time.Second * 2,
		RepeatHealthcheck:    time.Second * 1,
	}
	serviceHealthcheckTwo := domain.ServiceHealthcheck{
		StopChecks:           make(chan struct{}),
		PercentOfAlivedForUp: 50,
		Type:                 "icmp",
		Timeout:              time.Second * 3,
		RepeatHealthcheck:    time.Second * 2,
	}
	serviceHealthcheckThree := domain.ServiceHealthcheck{
		StopChecks:           make(chan struct{}),
		PercentOfAlivedForUp: 50,
		Type:                 "tcp",
		Timeout:              time.Second * 4,
		RepeatHealthcheck:    time.Second * 3,
	}
	return serviceHealthcheckOne, serviceHealthcheckTwo, serviceHealthcheckThree
}

func createApplicationServersForTests() ([]*domain.ApplicationServer, []*domain.ApplicationServer, *domain.ApplicationServer) {
	serverHealthcheckOne, serverHealthcheckTwo, serverHealthcheckThree, serverHealthcheckFour, serverHealthcheckFive := createServerHealthchecksForTests()
	appSrvOne := &domain.ApplicationServer{
		ServerIP:          "1.1.1.1",
		ServerPort:        "1111",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckOne,
	}
	appSrvTwo := &domain.ApplicationServer{
		ServerIP:          "2.2.2.2",
		ServerPort:        "2222",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckTwo,
	}
	appSrvThree := &domain.ApplicationServer{
		ServerIP:          "3.3.3.3",
		ServerPort:        "3333",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckThree,
	}
	appSrvFour := &domain.ApplicationServer{
		ServerIP:          "4.4.4.4",
		ServerPort:        "4444",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckFour,
	}
	appSrvFive := &domain.ApplicationServer{
		ServerIP:          "5.5.5.5",
		ServerPort:        "5555",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckFive,
	}
	appSrvSix := &domain.ApplicationServer{
		ServerIP:          "1.1.1.1",
		ServerPort:        "1111",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckOne,
	}
	return []*domain.ApplicationServer{appSrvOne, appSrvTwo, appSrvFive}, []*domain.ApplicationServer{appSrvThree, appSrvFour}, appSrvSix
}

func createServerHealthchecksForTests() (domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck) {
	advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo, advancedHealthcheckParametersThree, advancedHealthcheckParametersFour := createAdvancedHealthcheckParametersForTests()

	serverHealthcheckOne := domain.ServerHealthcheck{
		TypeOfCheck:                   "http-advanced",
		HealthcheckAddress:            "127.0.1.1:9000",
		AdvancedHealthcheckParameters: []domain.AdvancedHealthcheckParameters{advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo},
	}

	serverHealthcheckTwo := domain.ServerHealthcheck{
		TypeOfCheck:        "icmp",
		HealthcheckAddress: "127.0.2.1",
	}

	serverHealthcheckThree := domain.ServerHealthcheck{
		TypeOfCheck:                   "http-advanced",
		HealthcheckAddress:            "127.0.3.1:8000",
		AdvancedHealthcheckParameters: []domain.AdvancedHealthcheckParameters{advancedHealthcheckParametersThree, advancedHealthcheckParametersFour},
	}

	serverHealthcheckFour := domain.ServerHealthcheck{
		TypeOfCheck:        "tcp",
		HealthcheckAddress: "127.0.4.1",
	}

	serverHealthcheckFive := domain.ServerHealthcheck{
		TypeOfCheck:        "tcp",
		HealthcheckAddress: "127.0.5.1",
	}

	return serverHealthcheckOne, serverHealthcheckTwo, serverHealthcheckThree, serverHealthcheckFour, serverHealthcheckFive
}

func createAdvancedHealthcheckParametersForTests() (domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters) {
	advancedHealthcheckParametersOne := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  true,
		UserDefinedData: map[string]interface{}{"one": "oneValue", "two": 2},
	}
	advancedHealthcheckParametersTwo := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  false,
		UserDefinedData: map[string]interface{}{"three": "threeValue", "four": 4},
	}

	advancedHealthcheckParametersThree := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  true,
		UserDefinedData: map[string]interface{}{"five": "fiveValue", "six": 6},
	}
	advancedHealthcheckParametersFour := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  false,
		UserDefinedData: map[string]interface{}{"seven": "sevenValue", "eight": 8},
	}
	return advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo, advancedHealthcheckParametersThree, advancedHealthcheckParametersFour
}
