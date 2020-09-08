package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/domain"
)

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
