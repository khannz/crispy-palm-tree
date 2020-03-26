package portadapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	rawMiddleDataToKeepalivedDConfigs = `real_server REAL_SERVER_IP REAL_SERVER_PORT {
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

// 	rawRowForKeepalivedConfig = `include keepalived.d/services-configured/KEEPALIVED_D_CONFIG_FILE_NAME
// `
)

// KeepalivedCustomizer ...
type KeepalivedCustomizer struct {
	techInterface                     string
	fwmarkNumber                      string
	pathToKeepalivedConfig            string
	pathToKeepalivedDConfigConfigured string
	pathToKeepalivedDConfigEnabled    string
	logging                           *logrus.Logger
}

// NewKeepalivedCustomizer ...
func NewKeepalivedCustomizer(techInterface,
	fwmarkNumber,
	pathToKeepalivedConfig,
	pathToKeepalivedDConfigConfigured,
	pathToKeepalivedDConfigEnabled string,
	logging *logrus.Logger) *KeepalivedCustomizer {
	return &KeepalivedCustomizer{
		techInterface:                     techInterface,
		fwmarkNumber:                      fwmarkNumber,
		pathToKeepalivedConfig:            pathToKeepalivedConfig,
		pathToKeepalivedDConfigConfigured: pathToKeepalivedDConfigConfigured,
		pathToKeepalivedDConfigEnabled:    pathToKeepalivedDConfigEnabled,
		logging:                           logging,
	}
}

// CustomizeKeepalived ...
func (keepalivedCustomizer *KeepalivedCustomizer) CustomizeKeepalived(serviceIP,
	servicePort string,
	realServersData map[string]string,
	deployedEntities map[string][]string,
	newNWBRequestUUID string) (map[string][]string, error) {
	newKeepalivedDConfigFileName, err := keepalivedCustomizer.modifyKeepalivedConfigFiles(serviceIP, servicePort, realServersData, deployedEntities, newNWBRequestUUID)
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
	realServersData map[string]string,
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

	// err = keepalivedCustomizer.writeNextRowToKeepalivedConfig(newKeepalivedDConfigFileName) // TODO: remove that
	// if err != nil {
	// 	return "", fmt.Errorf("Error write next row to keepalived config: %v", err)
	// }

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

	for realServerIP, realServerPort := range realServersData {
		err = keepalivedCustomizer.writeMiddleDataForKeepalivedDConfigFile(serviceIP,
			servicePort,
			nextDummyNumberString,
			realServerIP,
			realServerPort,
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

// func (keepalivedCustomizer *KeepalivedCustomizer) writeNextRowToKeepalivedConfig(newKeepalivedDConfigFileName string) error {  // TODO: remove that
// 	rowForKeepalivedConfig := strings.ReplaceAll(rawRowForKeepalivedConfig, "KEEPALIVED_D_CONFIG_FILE_NAME", newKeepalivedDConfigFileName)
// 	keepalivedFile, err := os.OpenFile(keepalivedCustomizer.pathToKeepalivedConfig, os.O_APPEND|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return fmt.Errorf("Can't open keepalived config file: %v", err)
// 	}
// 	defer keepalivedFile.Close()
// 	if _, err = keepalivedFile.WriteString(rowForKeepalivedConfig); err != nil {
// 		return fmt.Errorf("Can't write data to keepalived config file: %v", err)
// 	}
// 	return nil
// }

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
	realServerIP,
	realServerPort,
	newFullKeepalivedDConfigFilePath string) error {

	rowsForKeepalivedDConfig := strings.ReplaceAll(rawMiddleDataToKeepalivedDConfigs, "REAL_SERVER_IP", realServerIP)
	rowsForKeepalivedDConfig = strings.ReplaceAll(rowsForKeepalivedDConfig, "REAL_SERVER_PORT", realServerPort)
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

// // RemoveRowFromKeepalivedConfigFile ... // TODO: remove that
// func (keepalivedCustomizer *KeepalivedCustomizer) RemoveRowFromKeepalivedConfigFile(rowInKeepalivedConfig, requestUUID string) error {
// 	err := removeRowFromFile(keepalivedCustomizer.pathToKeepalivedConfig, rowInKeepalivedConfig)
// 	if err != nil {
// 		return fmt.Errorf("can't remove row from keepalived config file: %v", err)
// 	}
// 	return nil
// }

// DetectKeepalivedConfigFiles ...
func (keepalivedCustomizer *KeepalivedCustomizer) DetectKeepalivedConfigFiles(serviceIP,
	servicePort string,
	deployedEntities map[string][]string,
	removeRequestUUID string) (map[string][]string, error) {
	keepalivedDConfigFileName, err := keepalivedCustomizer.getKeepalivedconfigFileName(serviceIP, servicePort) // rework!!!
	if err != nil {
		return deployedEntities, fmt.Errorf("can't  detectKeepalived config files: %v", err)
	}
	deployedEntities["newKeepalivedDConfigFileName"] = []string{keepalivedDConfigFileName}
	deployedEntities["newFullKeepalivedDConfigFilePath"] = []string{keepalivedCustomizer.pathToKeepalivedDConfigConfigured + keepalivedDConfigFileName}
	deployedEntities["fullPathToEnabledKeepalivedDFile"] = []string{keepalivedCustomizer.pathToKeepalivedDConfigEnabled + keepalivedDConfigFileName}
	return deployedEntities, nil
}

func (keepalivedCustomizer *KeepalivedCustomizer) getKeepalivedconfigFileName(serviceIP, servicePort string) (string, error) {
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

// func (keepalivedCustomizer *KeepalivedCustomizer) detectKeepalivedRowInConfigFile(serviceIP, servicePort string) (string, error) { // TODO: remove that
// 	file, err := os.OpenFile(keepalivedCustomizer.pathToKeepalivedConfig, os.O_RDWR, 0644)
// 	if err != nil {
// 		return "", fmt.Errorf("can't open file %v, got error %v", keepalivedCustomizer.pathToKeepalivedConfig, err)
// 	}
// 	defer file.Close()

// 	fileBytes, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		return "", fmt.Errorf("can't read data from file %v, got error %v", keepalivedCustomizer.pathToKeepalivedConfig, err)
// 	}

// 	keepalivedConfigData := string(fileBytes)
// 	keepalivedRowTowToSearch := "include keepalived.d/services-configured/(" + serviceIP + "-" + servicePort + "_dummy0-" + ".*.conf)"
// 	re := regexp.MustCompile(keepalivedRowTowToSearch)
// 	finded := re.FindAllStringSubmatch(keepalivedConfigData, -1)
// 	if len(finded) >= 1 {
// 		if len(finded[0]) >= 2 {
// 			return finded[0][1], nil
// 		}
// 	}
// 	return "", fmt.Errorf("can't find row for ip %v port %v in keepalived config file %v", serviceIP, servicePort, keepalivedCustomizer.pathToKeepalivedConfig)
// }

// ReloadKeepalived ...
func (keepalivedCustomizer *KeepalivedCustomizer) ReloadKeepalived(requestUUID string) error {
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
