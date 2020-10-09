package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-controller/domain"
)

func createServicesInfoForTests(applicattionServersOne []*domain.ApplicationServer, applicattionServersTwo []*domain.ApplicationServer) (*domain.ServiceInfo, *domain.ServiceInfo, *domain.ServiceInfo) {
	// serviceHealthcheckOne, serviceHealthcheckTwo, _ := createServicesHealthchecksForTests()
	serviceInfoOne := &domain.ServiceInfo{
		Address:               "111.111.111.111:111",
		IP:                    "111.111.111.111",
		Port:                  "111",
		IsUp:                  true,
		BalanceType:           "rr",
		RoutingType:           "tunneling",
		Protocol:              "tcp",
		AlivedAppServersForUp: 2,
		HCType:                "icmp",
		HCRepeat:              5 * time.Second,
		HCTimeout:             1 * time.Second,
		HCNearFieldsMode:      false,
		HCUserDefinedData:     make(map[string]string, 1),
		HCRetriesForUP:        5,
		HCRetriesForDown:      2,
		ApplicationServers:    applicattionServersOne,
		HCStop:                make(chan struct{}, 1),
		HCStopped:             make(chan struct{}, 1),
	}
	serviceInfoTwo := &domain.ServiceInfo{
		Address:               "222.222.222.222:222",
		IP:                    "222.222.222.222",
		Port:                  "222",
		IsUp:                  true,
		BalanceType:           "rr",
		RoutingType:           "tunneling",
		Protocol:              "tcp",
		AlivedAppServersForUp: 2,
		HCType:                "tcp",
		HCRepeat:              5 * time.Second,
		HCTimeout:             1 * time.Second,
		HCNearFieldsMode:      false,
		HCUserDefinedData:     make(map[string]string, 1),
		HCRetriesForUP:        5,
		HCRetriesForDown:      2,
		ApplicationServers:    applicattionServersTwo,
		HCStop:                make(chan struct{}, 1),
		HCStopped:             make(chan struct{}, 1),
	}
	serviceInfoThree := &domain.ServiceInfo{
		Address:     "111.111.111.111:111",
		IP:          "111.111.111.111",
		Port:        "111",
		BalanceType: "rr",
		RoutingType: "tunneling",
	}

	return serviceInfoOne, serviceInfoTwo, serviceInfoThree
}

// func createServicesHealthchecksForTests() (domain.ServiceHealthcheck, domain.ServiceHealthcheck, domain.ServiceHealthcheck) {
// 	serviceHealthcheckOne := domain.ServiceHealthcheck{
// 		StopChecks:           make(chan struct{}),
// 		PercentOfAlivedForUp: 50,
// 		Type:                 "http",
// 		Timeout:              time.Second * 2,
// 		// RepeatHealthcheck:    time.Second * 1,
// 	}
// 	serviceHealthcheckTwo := domain.ServiceHealthcheck{
// 		StopChecks:           make(chan struct{}),
// 		PercentOfAlivedForUp: 50,
// 		Type:                 "icmp",
// 		Timeout:              time.Second * 3,
// 		// RepeatHealthcheck:    time.Second * 2,
// 	}
// 	serviceHealthcheckThree := domain.ServiceHealthcheck{
// 		StopChecks:           make(chan struct{}),
// 		PercentOfAlivedForUp: 50,
// 		Type:                 "tcp",
// 		Timeout:              time.Second * 4,
// 		// RepeatHealthcheck:    time.Second * 3,
// 	}
// 	return serviceHealthcheckOne, serviceHealthcheckTwo, serviceHealthcheckThree
// }

func createApplicationServersForTests() ([]*domain.ApplicationServer, []*domain.ApplicationServer, *domain.ApplicationServer) {
	// serverHealthcheckOne, serverHealthcheckTwo, serverHealthcheckThree, serverHealthcheckFour, serverHealthcheckFive := createServerHealthchecksForTests()
	appSrvOne := &domain.ApplicationServer{
		IP:   "1.1.1.1",
		Port: "1111",
		IsUp: true,
		// ServerHealthcheck: serverHealthcheckOne,
	}
	appSrvTwo := &domain.ApplicationServer{
		IP:   "2.2.2.2",
		Port: "2222",
		IsUp: true,
		// ServerHealthcheck: serverHealthcheckTwo,
	}
	appSrvThree := &domain.ApplicationServer{
		IP:   "3.3.3.3",
		Port: "3333",
		IsUp: true,
		// ServerHealthcheck: serverHealthcheckThree,
	}
	appSrvFour := &domain.ApplicationServer{
		IP:   "4.4.4.4",
		Port: "4444",
		IsUp: true,
		// ServerHealthcheck: serverHealthcheckFour,
	}
	appSrvFive := &domain.ApplicationServer{
		IP:   "5.5.5.5",
		Port: "5555",
		IsUp: true,
		// ServerHealthcheck: serverHealthcheckFive,
	}
	appSrvSix := &domain.ApplicationServer{
		IP:   "1.1.1.1",
		Port: "1111",
		IsUp: true,
		// ServerHealthcheck: serverHealthcheckOne,
	}
	return []*domain.ApplicationServer{appSrvOne, appSrvTwo, appSrvFive}, []*domain.ApplicationServer{appSrvThree, appSrvFour}, appSrvSix
}

// func createServerHealthchecksForTests() (domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck) {
// 	advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo, advancedHealthcheckParametersThree, advancedHealthcheckParametersFour := createAdvancedHealthcheckParametersForTests()

// 	serverHealthcheckOne := domain.ServerHealthcheck{
// 		TypeOfCheck:                   "http-advanced",
// 		HealthcheckAddress:            "127.0.1.1:9000",
// 		AdvancedHealthcheckParameters: []domain.AdvancedHealthcheckParameters{advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo},
// 	}

// 	serverHealthcheckTwo := domain.ServerHealthcheck{
// 		TypeOfCheck:        "icmp",
// 		HealthcheckAddress: "127.0.2.1",
// 	}

// 	serverHealthcheckThree := domain.ServerHealthcheck{
// 		TypeOfCheck:                   "http-advanced",
// 		HealthcheckAddress:            "127.0.3.1:8000",
// 		AdvancedHealthcheckParameters: []domain.AdvancedHealthcheckParameters{advancedHealthcheckParametersThree, advancedHealthcheckParametersFour},
// 	}

// 	serverHealthcheckFour := domain.ServerHealthcheck{
// 		TypeOfCheck:        "tcp",
// 		HealthcheckAddress: "127.0.4.1",
// 	}

// 	serverHealthcheckFive := domain.ServerHealthcheck{
// 		TypeOfCheck:        "tcp",
// 		HealthcheckAddress: "127.0.5.1",
// 	}

// 	return serverHealthcheckOne, serverHealthcheckTwo, serverHealthcheckThree, serverHealthcheckFour, serverHealthcheckFive
// }

// func createAdvancedHealthcheckParametersForTests() (domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters) {
// 	advancedHealthcheckParametersOne := domain.AdvancedHealthcheckParameters{
// 		NearFieldsMode:  true,
// 		UserDefinedData: map[string]string{"one": "oneValue", "two": 2},
// 	}
// 	advancedHealthcheckParametersTwo := domain.AdvancedHealthcheckParameters{
// 		NearFieldsMode:  false,
// 		UserDefinedData: map[string]string{"three": "threeValue", "four": 4},
// 	}

// 	advancedHealthcheckParametersThree := domain.AdvancedHealthcheckParameters{
// 		NearFieldsMode:  true,
// 		UserDefinedData: map[string]string{"five": "fiveValue", "six": 6},
// 	}
// 	advancedHealthcheckParametersFour := domain.AdvancedHealthcheckParameters{
// 		NearFieldsMode:  false,
// 		UserDefinedData: map[string]string{"seven": "sevenValue", "eight": 8},
// 	}
// 	return advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo, advancedHealthcheckParametersThree, advancedHealthcheckParametersFour
// }

// MockIPVSWorker ...
type MockIPVSWorker struct{}

// NewService ...
func (mockIPVSWorker *MockIPVSWorker) NewService(string, uint16, uint32, string, uint16, map[string]uint16, string) error {
	return nil
}

// RemoveService ...
func (mockIPVSWorker *MockIPVSWorker) RemoveService(string, uint16, uint16, string) error {
	return nil
}

// AddApplicationServersForService ...
func (mockIPVSWorker *MockIPVSWorker) AddApplicationServersForService(string, uint16, uint32, string, uint16, map[string]uint16, string) error {
	return nil
}

// RemoveApplicationServersFromService ...
func (mockIPVSWorker *MockIPVSWorker) RemoveApplicationServersFromService(string, uint16, uint32, string, uint16, map[string]uint16, string) error {
	return nil
}

// Flush ...
func (mockIPVSWorker *MockIPVSWorker) Flush() error {
	return nil
}

// MockTunnelMaker ...
type MockTunnelMaker struct{}

// CreateTunnel ...
func (mockTunnelMaker *MockTunnelMaker) CreateTunnel(*domain.TunnelForApplicationServer, string) error {
	return nil
}

// CreateTunnels ...
func (mockTunnelMaker *MockTunnelMaker) CreateTunnels([]*domain.TunnelForApplicationServer, string) ([]*domain.TunnelForApplicationServer, error) {
	return nil, nil
}

// RemoveTunnel ...
func (mockTunnelMaker *MockTunnelMaker) RemoveTunnel(*domain.TunnelForApplicationServer, string) error {
	return nil
}

// RemoveTunnels ...
func (mockTunnelMaker *MockTunnelMaker) RemoveTunnels([]*domain.TunnelForApplicationServer, string) ([]*domain.TunnelForApplicationServer, error) {
	return nil, nil
}

// RemoveTunnels ...
func (mockTunnelMaker *MockTunnelMaker) RemoveAllTunnels([]*domain.TunnelForApplicationServer, string) error {
	return nil
}

// MockHCWorker ...
type MockHCWorker struct{}

func (mockHCWorker *MockHCWorker) StartHealthchecksForCurrentServices([]*domain.ServiceInfo) error {
	return nil
}
func (mockHCWorker *MockHCWorker) NewServiceToHealtchecks(*domain.ServiceInfo) error      { return nil }
func (mockHCWorker *MockHCWorker) RemoveServiceFromHealtchecks(*domain.ServiceInfo) error { return nil }
func (mockHCWorker *MockHCWorker) UpdateServiceAtHealtchecks(*domain.ServiceInfo) (*domain.ServiceInfo, error) {
	return nil, nil
}
func (mockHCWorker *MockHCWorker) GetServiceState(*domain.ServiceInfo) (*domain.ServiceInfo, error) {
	return nil, nil
}
func (mockHCWorker *MockHCWorker) GetServicesState() ([]*domain.ServiceInfo, error) { return nil, nil }
func (mockHCWorker *MockHCWorker) ConnectToHealtchecks() error                      { return nil }
func (mockHCWorker *MockHCWorker) DisconnectFromHealtchecks()                       {}

// MockCommandGenerator ...
type MockCommandGenerator struct{}

// GenerateCommandsForApplicationServers ...
func (MockCommandGenerator *MockCommandGenerator) GenerateCommandsForApplicationServers(*domain.ServiceInfo, string) error {
	return nil
}
