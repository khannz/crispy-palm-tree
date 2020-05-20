package portadapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/executor"
)

const tunnelFileMakerEntityName = "tunnel-file-maker"

const (
	rawDataForTunnelFile = `DEVICE=TUNNEL_NAME
ONBOOT=yes
TYPE=IPIP
PEER_OUTER_IPADDR=REAL_SERVER_IP
MTU=9000
`
	rawDataForTunnelRouteFile = `REAL_SERVER_IP/32 dev TUNNEL_NAME table srv_health_check`
	rowForSysctlConf          = `net.ipv4.conf.TUNNEL_NAME.rp_filter=0`
)

// TunnelFileMaker ...
type TunnelFileMaker struct {
	pathToIfcfgTunnelFiles string
	sysctlConfFilePath     string
	isMockMode             bool
	logging                *logrus.Logger
}

// NewTunnelFileMaker ...
func NewTunnelFileMaker(pathToIfcfgTunnelFiles string,
	sysctlConfFilePath string,
	isMockMode bool,
	logging *logrus.Logger) *TunnelFileMaker {
	return &TunnelFileMaker{
		pathToIfcfgTunnelFiles: pathToIfcfgTunnelFiles,
		sysctlConfFilePath:     sysctlConfFilePath,
		isMockMode:             isMockMode,
		logging:                logging,
	}
}

// CreateTunnels ...
func (tunnelFileMaker *TunnelFileMaker) CreateTunnels(tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	createTunnelUUID string) ([]*domain.TunnelForApplicationServer, error) {
	newTunnelsFilesInfo := []*domain.TunnelForApplicationServer{}
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		if tunnelFilesInfo.ServicesToTunnelCount == 0 {
			if err := tunnelFileMaker.CreateTunnel(tunnelFilesInfo, createTunnelUUID); err != nil {
				return nil, fmt.Errorf("can't create tunnel: %v", err)
			}
		}
		newTunnelFilesInfo := &domain.TunnelForApplicationServer{
			ApplicationServerIP:   tunnelFilesInfo.ApplicationServerIP,
			IfcfgTunnelFile:       tunnelFilesInfo.IfcfgTunnelFile,
			RouteTunnelFile:       tunnelFilesInfo.RouteTunnelFile,
			SysctlConfFile:        tunnelFilesInfo.SysctlConfFile,
			TunnelName:            tunnelFilesInfo.TunnelName,
			ServicesToTunnelCount: tunnelFilesInfo.ServicesToTunnelCount + 1,
		}
		newTunnelsFilesInfo = append(newTunnelsFilesInfo, newTunnelFilesInfo)
	}
	return newTunnelsFilesInfo, nil
}

// CreateTunnel ...
func (tunnelFileMaker *TunnelFileMaker) CreateTunnel(tunnelFilesInfo *domain.TunnelForApplicationServer,
	createTunnelUUID string) error {
	newTunnelName, err := tunnelFileMaker.chooseNewTunnelName()
	if err != nil {
		return fmt.Errorf("can't choose new tunnel name: %v", err)
	}
	sNewTunnelName := strconv.Itoa(newTunnelName)
	newIfcfgTunnelFileFullPath := tunnelFileMaker.pathToIfcfgTunnelFiles + "ifcfg-" + "tun" + sNewTunnelName
	newRouteTunnelFileFullPath := tunnelFileMaker.pathToIfcfgTunnelFiles + "route-" + sNewTunnelName
	newSysctlConfFileFullPath := tunnelFileMaker.sysctlConfFilePath + sNewTunnelName + "-sysctl.conf"

	tunnelFilesInfo.TunnelName = sNewTunnelName
	tunnelFilesInfo.IfcfgTunnelFile = newIfcfgTunnelFileFullPath
	tunnelFilesInfo.RouteTunnelFile = newRouteTunnelFileFullPath
	tunnelFilesInfo.SysctlConfFile = newSysctlConfFileFullPath

	if err := tunnelFileMaker.writeNewTunnelFile(tunnelFilesInfo,
		createTunnelUUID); err != nil {
		return fmt.Errorf("can't write new tunnel files: %v", err)
	}

	if err := tunnelFileMaker.ExecuteCommandForTunnel("tun"+sNewTunnelName, "up", createTunnelUUID); err != nil {
		return fmt.Errorf("can't execute command for up tunnel: %v", err)
	}

	return nil
}

func (tunnelFileMaker *TunnelFileMaker) chooseNewTunnelName() (int, error) { // TODO: rework that, it's too hard just for new number
	files, err := ioutil.ReadDir(tunnelFileMaker.pathToIfcfgTunnelFiles)
	if err != nil {
		return 0, fmt.Errorf("read dir %v, got error %v", tunnelFileMaker.pathToIfcfgTunnelFiles, err)
	}

	var sliceOfOldTunelNames []string
	for _, f := range files {
		if strings.Contains(f.Name(), "ifcfg-tun") {
			sliceOfOldTunelNames = append(sliceOfOldTunelNames, strings.TrimPrefix(f.Name(), "ifcfg-tun"))
		}
	}

	var nextTunnelName int
	if len(sliceOfOldTunelNames) > 0 { // TODO: take last "free"
		sort.Sort(sort.Reverse(sort.StringSlice(sliceOfOldTunelNames)))
		nextTunnelName, err = strconv.Atoi(sliceOfOldTunelNames[0])
		if err != nil {
			return 0, fmt.Errorf("can't convert slice of old tunel names to string, got error: %v", err)
		}
		nextTunnelName++
	}
	return nextTunnelName, nil
}

func (tunnelFileMaker *TunnelFileMaker) writeNewTunnelFile(tunnelFilesInfo *domain.TunnelForApplicationServer,
	createTunnelUUID string) error {
	newDataForTunnelFile := strings.ReplaceAll(rawDataForTunnelFile,
		"TUNNEL_NAME",
		"tun"+tunnelFilesInfo.TunnelName)
	newDataForTunnelFile = strings.ReplaceAll(newDataForTunnelFile,
		"REAL_SERVER_IP",
		tunnelFilesInfo.ApplicationServerIP)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("creating new ifcfg for tunnel file: %v", tunnelFilesInfo.IfcfgTunnelFile)
	err := ioutil.WriteFile(tunnelFilesInfo.IfcfgTunnelFile, []byte(newDataForTunnelFile+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("can't write new tunnekt to file %v, got error: %v",
			tunnelFilesInfo.IfcfgTunnelFile, err)
	}

	dataForTunnelRouteFile := strings.ReplaceAll(rawDataForTunnelRouteFile,
		"REAL_SERVER_IP", tunnelFilesInfo.ApplicationServerIP)
	dataForTunnelRouteFile = strings.ReplaceAll(dataForTunnelRouteFile,
		"TUNNEL_NAME",
		"tun"+tunnelFilesInfo.TunnelName)

	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("creating new route for tunnel file: %v", tunnelFilesInfo.RouteTunnelFile)
	err = ioutil.WriteFile(tunnelFilesInfo.RouteTunnelFile,
		[]byte(dataForTunnelRouteFile+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("can't write new tunnell route to file %v, got error: %v",
			tunnelFileMaker.pathToIfcfgTunnelFiles+"ifcfg-"+tunnelFilesInfo.TunnelName, err)
	}

	newRowForSysctlConf := strings.ReplaceAll(rowForSysctlConf,
		"TUNNEL_NAME",
		"tun"+tunnelFilesInfo.TunnelName)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("new sysctl config file name: %v", tunnelFilesInfo.SysctlConfFile)
	err = ioutil.WriteFile(tunnelFilesInfo.SysctlConfFile, []byte(newRowForSysctlConf+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("can't write sysctl conf %v, got error: %v",
			tunnelFilesInfo.SysctlConfFile, err)
	}
	return nil
}

// ExecuteCommandForTunnel ...
func (tunnelFileMaker *TunnelFileMaker) ExecuteCommandForTunnel(tunnelName string,
	arg string,
	createTunnelUUID string) error {
	if tunnelFileMaker.isMockMode {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": createTunnelUUID,
		}).Infof("mock of execute ip command for tunnels %v", tunnelName)
		return nil
	}
	args := []string{"link", "set", "dev", tunnelName, arg}
	stdout, stderr, exitCode, err := executor.Execute("ip", "", args)
	if err != nil {
		return fmt.Errorf("when execute command %v, got error: %v", "ip link set dev "+tunnelName+" "+arg, err)
	}
	if exitCode != 0 {
		return fmt.Errorf("when execute command %v, got exit code != 0: stdout: %v, stderr: %v, exitCode: %v",
			"ip link set dev "+tunnelName+" "+arg,
			string(stdout),
			string(stderr),
			string(exitCode))
	}
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Debugf("result of execute ip command: stdout: %v, stderr: %v, exitCode: %v", string(stdout), string(stderr), string(exitCode))

	return nil
}

// RemoveTunnels ...
func (tunnelFileMaker *TunnelFileMaker) RemoveTunnels(tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	removeTunnelUUID string) ([]*domain.TunnelForApplicationServer, error) {
	newTunnelsFilesInfo := []*domain.TunnelForApplicationServer{}
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": removeTunnelUUID,
		}).Debugf("remove tunnel %v; files: %v; %v; %v", tunnelFilesInfo.ApplicationServerIP, tunnelFilesInfo.IfcfgTunnelFile, tunnelFilesInfo.RouteTunnelFile, tunnelFilesInfo.SysctlConfFile)

		if tunnelFilesInfo.ServicesToTunnelCount == 1 {
			if err := tunnelFileMaker.RemoveTunnel(tunnelFilesInfo, removeTunnelUUID); err != nil {
				return nil, fmt.Errorf("can't remove tunnel files: %v", err)
			}
		}
		tunnelFilesInfo.ServicesToTunnelCount = tunnelFilesInfo.ServicesToTunnelCount - 1
		newTunnelFilesInfo := *tunnelFilesInfo
		newTunnelsFilesInfo = append(newTunnelsFilesInfo, &newTunnelFilesInfo)
	}
	return newTunnelsFilesInfo, nil
}

// RemoveTunnel ...
func (tunnelFileMaker *TunnelFileMaker) RemoveTunnel(tunnelFilesInfo *domain.TunnelForApplicationServer,
	removeTunnelUUID string) error {
	var err error
	if err = tunnelFileMaker.removeFile(tunnelFilesInfo.IfcfgTunnelFile, removeTunnelUUID); err != nil {
		return err
	}
	if err = tunnelFileMaker.removeFile(tunnelFilesInfo.RouteTunnelFile, removeTunnelUUID); err != nil {
		return err
	}
	if err = tunnelFileMaker.removeFile(tunnelFilesInfo.SysctlConfFile, removeTunnelUUID); err != nil {
		return err
	}

	if err := tunnelFileMaker.ExecuteCommandForTunnel("tun"+tunnelFilesInfo.TunnelName, "down", removeTunnelUUID); err != nil {
		return fmt.Errorf("can't execute command for down tunnel: %v", err)
	}

	return nil
}

func (tunnelFileMaker *TunnelFileMaker) removeFile(filePath, requestUUID string) error {
	if err := os.Remove(filePath); err != nil {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": requestUUID,
		}).Errorf("can't remove tunnel file, got error: %v", filePath, err)
		return err
	}
	return nil
}
