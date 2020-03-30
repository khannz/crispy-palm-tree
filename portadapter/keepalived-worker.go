package portadapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/executor"
)

const keepalivedCustomizerEntityName = "keepalived-customizer"

const (
	rawStartDataToKeepalivedDConfigs = `virtual_server SERVICE_IP SERVICE_PORT {
lvs_method TUN         # L3 DSR
lvs_sched mh           # src addr hash
sh-fallback
persistence_timeout    900
persistence_granularity 255.255.255.255
protocol TCP
delay_loop 2           # healthcheck interval
alpha
omega

quorum_up "/sbin/ifconfig dummy0:DUMMY_NUMBER SERVICE_IP netmask 255.255.255.255 -arp up" root
quorum_down "/sbin/ifconfig dummy0:DUMMY_NUMBER down" root

`
	rawMiddleDataToKeepalivedDConfigs = `real_server APPLICATION_SERVER_IP APPLICATION_SERVER_PORT {
  TCP_CHECK {
   connect_timeout 2
   retry 3
   bindto TECH_INTERFACE
   fwmark FWMARK_NUMBER
  }
}
`
	rawEndDataToKeepalivedDConfigs = `}
`
	rawNewKeepalivedDConfigFileName = `SERVICE_IP-SERVICE_PORT_dummy0-DUMMY_NUMBER.conf`

	rawRowForKeepalivedConfig = `include keepalived.d/services-configured/KEEPALIVED_D_CONFIG_FILE_NAME
`
)

// KeepalivedCustomizer ...
type KeepalivedCustomizer struct {
	techInterface                     string
	fwmarkNumber                      string
	pathToKeepalivedConfig            string
	pathToKeepalivedDConfigConfigured string
	pathToKeepalivedDConfigEnabled    string
	isMockMode                        bool
	logging                           *logrus.Logger
}

// NewKeepalivedCustomizer ...
func NewKeepalivedCustomizer(techInterface,
	fwmarkNumber,
	pathToKeepalivedConfig,
	pathToKeepalivedDConfigConfigured,
	pathToKeepalivedDConfigEnabled string,
	isMockMode bool,
	logging *logrus.Logger) *KeepalivedCustomizer {
	return &KeepalivedCustomizer{
		techInterface:                     techInterface,
		fwmarkNumber:                      fwmarkNumber,
		pathToKeepalivedConfig:            pathToKeepalivedConfig,
		pathToKeepalivedDConfigConfigured: pathToKeepalivedDConfigConfigured,
		pathToKeepalivedDConfigEnabled:    pathToKeepalivedDConfigEnabled,
		isMockMode:                        isMockMode,
		logging:                           logging,
	}
}

// CustomizeKeepalived ...
func (keepalivedCustomizer *KeepalivedCustomizer) CustomizeKeepalived(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	deployedEntities map[string][]string,
	newNWBRequestUUID string) (map[string][]string, error) {
	newKeepalivedDConfigFileName, err := keepalivedCustomizer.modifyKeepalivedConfigFiles(serviceIP, servicePort, applicationServers, deployedEntities, newNWBRequestUUID)
	if err != nil {
		return deployedEntities, fmt.Errorf("Error when modify keepalived config: %v", err)
	}

	fullPathToEnabledKeepalivedDFile, err := keepalivedCustomizer.makeSymlinkForKeepalived(newKeepalivedDConfigFileName)
	if err != nil {
		return deployedEntities, fmt.Errorf("Error try create symlink for keepalived: %v", err)
	}
	deployedEntities["fullPathToEnabledKeepalivedDFile"] = []string{fullPathToEnabledKeepalivedDFile}

	err = keepalivedCustomizer.ReloadKeepalived(newNWBRequestUUID)
	if err != nil {
		return deployedEntities, fmt.Errorf("can't reload keepalived: %v", err)
	}

	return deployedEntities, nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) modifyKeepalivedConfigFiles(serviceIP,
	servicePort string,
	applicationServers map[string]string,
	deployedEntities map[string][]string,
	newNWBRequestUUID string) (string, error) {
	nextDummyNumber, err := keepalivedCustomizer.getNextDummyNumber()
	if err != nil {
		return "", fmt.Errorf("Error get next dummy number: %v", err)
	}

	nextDummyNumberString := strconv.Itoa(nextDummyNumber)
	newKeepalivedDConfigFileName := keepalivedCustomizer.makeNewKeepalivedDConfigFileName(serviceIP,
		servicePort,
		nextDummyNumberString)

	newRowInKeepalivedConfig, err := keepalivedCustomizer.writeNextRowToKeepalivedConfig(newKeepalivedDConfigFileName) // TODO: remove that
	if err != nil {
		return "", fmt.Errorf("Error write next row to keepalived config: %v", err)
	}
	deployedEntities["rowInKeepalivedConfig"] = []string{newRowInKeepalivedConfig}
	deployedEntities["newKeepalivedDConfigFileName"] = []string{newKeepalivedDConfigFileName}

	newFullKeepalivedDConfigFilePath := keepalivedCustomizer.pathToKeepalivedDConfigConfigured + newKeepalivedDConfigFileName

	err = keepalivedCustomizer.writeStartDataForKeepalivedDConfigFile(serviceIP,
		servicePort,
		nextDummyNumberString,
		newFullKeepalivedDConfigFilePath)
	if err != nil {
		return "", fmt.Errorf("Error create new keepalived d config file: %v", err)
	}
	deployedEntities["newFullKeepalivedDConfigFilePath"] = []string{newFullKeepalivedDConfigFilePath}

	for applicationServerIP, applicationServerPort := range applicationServers {
		err = keepalivedCustomizer.writeMiddleDataForKeepalivedDConfigFile(serviceIP,
			servicePort,
			nextDummyNumberString,
			applicationServerIP,
			applicationServerPort,
			newFullKeepalivedDConfigFilePath)
		if err != nil {
			return "", fmt.Errorf("Error create new keepalived d config file: %v", err)
		}
	}
	err = keepalivedCustomizer.writeEndDataForKeepalivedDConfigFile(newFullKeepalivedDConfigFilePath)
	if err != nil {
		return "", fmt.Errorf("Error create new keepalived d config file: %v", err)
	}
	return newKeepalivedDConfigFileName, nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) getNextDummyNumber() (int, error) {
	keepalivedData, err := ioutil.ReadFile(keepalivedCustomizer.pathToKeepalivedConfig)
	if err != nil {
		return 0, fmt.Errorf("Read file error: %v", err)
	}
	re := regexp.MustCompile(`dummy0-(.*).conf`)
	oldDummys := re.FindAllStringSubmatch(string(keepalivedData), -1)
	var dummyNumber int
	if len(oldDummys) > 0 {
		lastDummyWithGroup := oldDummys[len(oldDummys)-1]
		dummyNumber, err = strconv.Atoi(lastDummyWithGroup[1])
		if err != nil {
			return 0, fmt.Errorf("Can't conver dummy number to int: %v", err)
		}
	}
	dummyNumber++
	return dummyNumber, nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) makeNewKeepalivedDConfigFileName(serviceIP,
	servicePort,
	nextDummyNumber string) string {
	newKeepalivedDConfigFileName := strings.ReplaceAll(rawNewKeepalivedDConfigFileName, "SERVICE_IP", serviceIP)
	newKeepalivedDConfigFileName = strings.ReplaceAll(newKeepalivedDConfigFileName, "SERVICE_PORT", servicePort)
	newKeepalivedDConfigFileName = strings.ReplaceAll(newKeepalivedDConfigFileName, "DUMMY_NUMBER", nextDummyNumber)
	return newKeepalivedDConfigFileName
}

func (keepalivedCustomizer *KeepalivedCustomizer) writeNextRowToKeepalivedConfig(newKeepalivedDConfigFileName string) (string, error) { // TODO: remove that
	rowForKeepalivedConfig := strings.ReplaceAll(rawRowForKeepalivedConfig, "KEEPALIVED_D_CONFIG_FILE_NAME", newKeepalivedDConfigFileName)
	keepalivedFile, err := os.OpenFile(keepalivedCustomizer.pathToKeepalivedConfig, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("Can't open keepalived config file: %v", err)
	}
	defer keepalivedFile.Close()
	if _, err = keepalivedFile.WriteString(rowForKeepalivedConfig); err != nil {
		return "", fmt.Errorf("Can't write data to keepalived config file: %v", err)
	}
	return rowForKeepalivedConfig, nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) writeStartDataForKeepalivedDConfigFile(serviceIP,
	servicePort,
	nextDummyNumberString,
	newFullKeepalivedDConfigFilePath string) error {
	rowsForStartKeepalivedDConfig := strings.ReplaceAll(rawStartDataToKeepalivedDConfigs, "SERVICE_IP", serviceIP)
	rowsForStartKeepalivedDConfig = strings.ReplaceAll(rowsForStartKeepalivedDConfig, "SERVICE_PORT", servicePort)
	rowsForStartKeepalivedDConfig = strings.ReplaceAll(rowsForStartKeepalivedDConfig, "DUMMY_NUMBER", nextDummyNumberString)

	keepalivedDFile, err := os.OpenFile(newFullKeepalivedDConfigFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Can't create keepalived.d config file: %v", err)
	}
	defer keepalivedDFile.Close()
	if _, err = keepalivedDFile.WriteString(rowsForStartKeepalivedDConfig); err != nil {
		return fmt.Errorf("Can't write start data to keepalived.d config file: %v", err)
	}
	return nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) writeMiddleDataForKeepalivedDConfigFile(serviceIP,
	servicePort,
	nextDummyNumberString,
	applicationServerIP,
	applicationServerPort,
	newFullKeepalivedDConfigFilePath string) error {

	rowsForKeepalivedDConfig := strings.ReplaceAll(rawMiddleDataToKeepalivedDConfigs, "APPLICATION_SERVER_IP", applicationServerIP)
	rowsForKeepalivedDConfig = strings.ReplaceAll(rowsForKeepalivedDConfig, "APPLICATION_SERVER_PORT", applicationServerPort)
	rowsForKeepalivedDConfig = strings.ReplaceAll(rowsForKeepalivedDConfig, "TECH_INTERFACE", keepalivedCustomizer.techInterface)
	rowsForKeepalivedDConfig = strings.ReplaceAll(rowsForKeepalivedDConfig, "FWMARK_NUMBER", keepalivedCustomizer.fwmarkNumber)

	keepalivedDFile, err := os.OpenFile(newFullKeepalivedDConfigFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Can't open keepalived.d config file: %v", err)
	}
	defer keepalivedDFile.Close()
	if _, err = keepalivedDFile.WriteString(rowsForKeepalivedDConfig); err != nil {
		return fmt.Errorf("Can't write middle data to keepalived.d config file: %v", err)
	}
	return nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) writeEndDataForKeepalivedDConfigFile(newFullKeepalivedDConfigFilePath string) error {
	keepalivedDFile, err := os.OpenFile(newFullKeepalivedDConfigFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Can't open keepalived.d config file: %v", err)
	}
	defer keepalivedDFile.Close()
	if _, err = keepalivedDFile.WriteString(rawEndDataToKeepalivedDConfigs); err != nil {
		return fmt.Errorf("Can't write end data to keepalived.d config file: %v", err)
	}
	return nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) makeSymlinkForKeepalived(newKeepalivedDConfigFileName string) (string, error) {
	fullPathToConfiguredKeepalivedDFile := keepalivedCustomizer.pathToKeepalivedDConfigConfigured + newKeepalivedDConfigFileName
	fullPathToEnabledKeepalivedDFile := keepalivedCustomizer.pathToKeepalivedDConfigEnabled + newKeepalivedDConfigFileName
	err := os.Symlink(fullPathToConfiguredKeepalivedDFile, fullPathToEnabledKeepalivedDFile)
	if err != nil {
		return "", err
	}
	return fullPathToEnabledKeepalivedDFile, nil
}

// GetApplicationServersByServiceIPAndPort ...
func (keepalivedCustomizer *KeepalivedCustomizer) GetApplicationServersByServiceIPAndPort(serviceIP, servicePort string, requestUUID string) ([]domain.ApplicationServer, error) {
	keepalivedDConfigFileName, err := keepalivedCustomizer.getKeepalivedConfigFileName(serviceIP, servicePort)
	if err != nil {
		return nil, fmt.Errorf("can't get get keepalived config file name: %v", err)
	}

	rawKeepalivedDConfigFileData, err := ioutil.ReadFile(keepalivedCustomizer.pathToKeepalivedDConfigConfigured + keepalivedDConfigFileName)
	if err != nil {
		return nil, fmt.Errorf("can't read file %v, got error %v", keepalivedCustomizer.pathToKeepalivedDConfigConfigured+keepalivedDConfigFileName, err)
	}

	applicationServers, err := getApplicationServers(string(rawKeepalivedDConfigFileData))
	if err != nil {
		return nil, fmt.Errorf("can't get application servers in  %v, got error %v", keepalivedCustomizer.pathToKeepalivedDConfigConfigured+keepalivedDConfigFileName, err)
	}

	return applicationServers, nil
}

// RemoveKeepalivedDConfigFile ...
func (keepalivedCustomizer *KeepalivedCustomizer) RemoveKeepalivedDConfigFile(keepalivedDConfigFile, requestUUID string) error {
	err := os.Remove(keepalivedDConfigFile)
	if err != nil {
		return fmt.Errorf("can't remove keepalived file %v, got error: %v", keepalivedDConfigFile, err)
	}
	return nil
}

// RemoveKeepalivedSymlink ...
func (keepalivedCustomizer *KeepalivedCustomizer) RemoveKeepalivedSymlink(symlinkFilePath, requestUUID string) error {
	err := os.Remove(symlinkFilePath)
	if err != nil {
		return fmt.Errorf("can't remove keepalived file %v, got error: %v", symlinkFilePath, err)
	}
	return nil
}

// DetectKeepalivedConfigFileRows ...
func (keepalivedCustomizer *KeepalivedCustomizer) DetectKeepalivedConfigFileRows(serviceIP,
	servicePort string,
	deployedEntities map[string][]string,
	requestUUID string) (map[string][]string, error) {
	regexForFindRowInkeepalivedConfigForSearch := `include keepalived.d\/services-configured\/(` + serviceIP + `-` + servicePort + `_dummy0-.*.conf)`
	re := regexp.MustCompile(regexForFindRowInkeepalivedConfigForSearch)

	rawFileData, err := ioutil.ReadFile(keepalivedCustomizer.pathToKeepalivedConfig)
	if err != nil {
		return deployedEntities, err
	}

	finded := re.FindAllStringSubmatch(string(rawFileData), -1)
	if len(finded) == 1 {
		lastWithGroup := finded[len(finded)-1]
		deployedEntities["rowInKeepalivedConfig"] = []string{lastWithGroup[0]}
		return deployedEntities, nil
	}
	return deployedEntities, fmt.Errorf("expect find one row in keepalived config file, but finded %v", len(finded))
}

// RemoveRowFromKeepalivedConfigFile ...
func (keepalivedCustomizer *KeepalivedCustomizer) RemoveRowFromKeepalivedConfigFile(rowInKeepalivedConfig, requestUUID string) error {
	err := removeRowFromFile(keepalivedCustomizer.pathToKeepalivedConfig, rowInKeepalivedConfig)
	if err != nil {
		return fmt.Errorf("can't remove row from keepalived config file: %v", err)
	}
	return nil
}

// DetectKeepalivedConfigFiles ...
func (keepalivedCustomizer *KeepalivedCustomizer) DetectKeepalivedConfigFiles(serviceIP,
	servicePort string,
	deployedEntities map[string][]string,
	removeRequestUUID string) (map[string][]string, error) {
	keepalivedDConfigFileName, err := keepalivedCustomizer.getKeepalivedConfigFileName(serviceIP, servicePort) // rework!!!
	if err != nil {
		return deployedEntities, fmt.Errorf("can't  detectKeepalived config files: %v", err)
	}
	deployedEntities["newKeepalivedDConfigFileName"] = []string{keepalivedDConfigFileName}
	deployedEntities["newFullKeepalivedDConfigFilePath"] = []string{keepalivedCustomizer.pathToKeepalivedDConfigConfigured + keepalivedDConfigFileName}
	deployedEntities["fullPathToEnabledKeepalivedDFile"] = []string{keepalivedCustomizer.pathToKeepalivedDConfigEnabled + keepalivedDConfigFileName}
	return deployedEntities, nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) getKeepalivedConfigFileName(serviceIP, servicePort string) (string, error) {
	files, err := ioutil.ReadDir(keepalivedCustomizer.pathToKeepalivedDConfigConfigured)
	if err != nil {
		return "", fmt.Errorf("read dir %v, got error %v", keepalivedCustomizer.pathToKeepalivedDConfigConfigured, err)
	}

	for _, fileName := range files {
		if strings.Contains(fileName.Name(), serviceIP+"-"+servicePort+"_dummy0-") {
			return fileName.Name(), nil
		}
	}

	return "", fmt.Errorf("in folder %v didn't find file names like %v-%v", keepalivedCustomizer.pathToKeepalivedDConfigConfigured, serviceIP, servicePort)
}

// ReloadKeepalived ...
func (keepalivedCustomizer *KeepalivedCustomizer) ReloadKeepalived(requestUUID string) error {
	if keepalivedCustomizer.isMockMode {
		keepalivedCustomizer.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": requestUUID,
		}).Info("mock of execute reload for keepalived")
		return nil
	}
	args := []string{"reload", "keepalived"}
	stdout, stderr, exitCode, err := executor.Execute("/usr/bin/systemctl", "", args)
	if err != nil {
		return fmt.Errorf("when execute command %v, got error: %v", "/usr/bin/systemctl reload keepalived", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("when execute command %v, got exit code != 0: stdout: %v, stderr: %v, exitCode: %v",
			"/usr/bin/systemctl reload keepalived",
			string(stdout),
			string(stderr),
			exitCode)
	}
	keepalivedCustomizer.logging.WithFields(logrus.Fields{
		"entity":     keepalivedCustomizerEntityName,
		"event uuid": requestUUID,
	}).Debugf("result of execute systemctl reload command: stdout: %v, stderr: %v, exitCode: %v", string(stdout), string(stderr), exitCode)
	return nil
}

// GetInfoAboutAllNWBServices ...
func (keepalivedCustomizer *KeepalivedCustomizer) GetInfoAboutAllNWBServices(requestUUID string) ([]domain.ServiceInfo, error) {
	files, err := ioutil.ReadDir(keepalivedCustomizer.pathToKeepalivedDConfigConfigured)
	if err != nil {
		return nil, fmt.Errorf("read dir %v, got error %v", keepalivedCustomizer.pathToKeepalivedDConfigConfigured, err)
	}
	errors := []error{}
	servicesInfo := []domain.ServiceInfo{}
	for _, fileName := range files {
		dataBytes, err := ioutil.ReadFile(keepalivedCustomizer.pathToKeepalivedDConfigConfigured + fileName.Name())
		if err != nil {
			errors = append(errors, fmt.Errorf("can't read file %v, got error %v", keepalivedCustomizer.pathToKeepalivedDConfigConfigured+fileName.Name(), err))
			continue
		}
		serviceInfo, err := getServiceInfoByKeepalivedConfigDFileData(string(dataBytes), requestUUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("can't get service info: %v", err))
			continue
		}
		servicesInfo = append(servicesInfo, *serviceInfo)
	}
	return servicesInfo, combineErrors(errors)
}

func getServiceInfoByKeepalivedConfigDFileData(fileData, requestUUID string) (*domain.ServiceInfo, error) {
	serviceIP, servicePort, err := getServiceIPAndPort(fileData)
	if err != nil {
		return nil, fmt.Errorf("can't get service ip and port: %v", err)
	}

	applicationServers, err := getApplicationServers(fileData)
	if err != nil {
		return nil, fmt.Errorf("can't get application servers data: %v", err)
	}

	healthcheckType, err := getHealthcheckType(fileData)
	if err != nil {
		return nil, fmt.Errorf("can't get healthcheckType: %v", err)
	}

	serviceInfo := &domain.ServiceInfo{
		ServiceIP:          serviceIP,
		ServicePort:        servicePort,
		ApplicationServers: applicationServers,
		HealthcheckType:    healthcheckType,
	}

	return serviceInfo, nil
}

func getServiceIPAndPort(fileData string) (string, string, error) {
	re := regexp.MustCompile(`virtual_server (.*.) (.*.) {`)
	finded := re.FindAllStringSubmatch(fileData, -1)
	if len(finded) >= 1 {
		allWithGroups := finded[len(finded)-1]
		if len(allWithGroups) >= 3 {
			return allWithGroups[1], allWithGroups[2], nil
		}
	}
	return "", "", fmt.Errorf("can't find service ip and port, finded elements: %v", len(finded))
}

func getApplicationServers(fileData string) ([]domain.ApplicationServer, error) {
	applicationServers := []domain.ApplicationServer{}

	re := regexp.MustCompile(`real_server (.*.) (.*.) {`)
	finded := re.FindAllStringSubmatch(fileData, -1)
	if len(finded) >= 1 {
		for _, allWithGroups := range finded {
			if !(len(allWithGroups) >= 3) {
				return applicationServers, fmt.Errorf("can't find application server ip and port, regexp finded elements: %v of 3 needed", len(allWithGroups))
			}
			applicationServer := getapplicationServerIPAndPortByRegexpData(allWithGroups)
			applicationServers = append(applicationServers, applicationServer)
		}
	} else {
		return applicationServers, fmt.Errorf("can't find application server ip and port, finded elements: %v", len(finded))
	}

	return applicationServers, nil
}

func getapplicationServerIPAndPortByRegexpData(regexpData []string) domain.ApplicationServer {
	applicationServerIP := regexpData[1]
	applicationServerPort := regexpData[2]
	return domain.ApplicationServer{
		ServerIP:   applicationServerIP,
		ServerPort: applicationServerPort,
	}
}

func getHealthcheckType(fileData string) (string, error) {
	re := regexp.MustCompile(` {\n  (.*.) {\n`)
	finded := re.FindAllStringSubmatch(fileData, -1)
	if len(finded) >= 1 {
		allWithGroups := finded[len(finded)-1]
		healthcheckType, err := chooseHealtcheckType(allWithGroups[1])
		if err != nil {
			return "", fmt.Errorf("can't choose healtcheck type: %v", err)
		}
		return healthcheckType, nil
	}
	return "", fmt.Errorf("can't healthcheck data, finded elements: %v", len(finded))
}

func chooseHealtcheckType(rawHealthcheck string) (string, error) {
	switch rawHealthcheck {
	case "TCP_CHECK":
		return "tcp", nil
	default:
		return "", fmt.Errorf("unknown healtchecktype: %v", rawHealthcheck)
	}
}
