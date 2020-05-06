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

// EnrichApplicationServersInfo add tunnel info to application servers struct
func (tunnelFileMaker *TunnelFileMaker) EnrichApplicationServersInfo(applicationServers []*domain.ApplicationServer,
	requestUUID string) ([]*domain.ApplicationServer, error) {
	enrichedApplicationServers := []*domain.ApplicationServer{}
	newTunnelName, err := tunnelFileMaker.chooseNewTunnelName()
	if err != nil {
		return nil, fmt.Errorf("can't choose new tunnel name: %v", err)
	}
	for _, applicationServer := range applicationServers {
		enrichedApplicationServer, err := tunnelFileMaker.EnrichApplicationServerInfo(applicationServer, newTunnelName, requestUUID)
		if err != nil {
			return enrichedApplicationServers, fmt.Errorf("can't enrich application server info: %v", err)
		}
		enrichedApplicationServers = append(enrichedApplicationServers, enrichedApplicationServer)
		newTunnelName++
	}
	return enrichedApplicationServers, nil
}

// EnrichApplicationServerInfo add tunnel info to application servers struct
func (tunnelFileMaker *TunnelFileMaker) EnrichApplicationServerInfo(applicationServers *domain.ApplicationServer,
	newTunnelName int,
	requestUUID string) (*domain.ApplicationServer, error) {
	sNewTunnelName := strconv.Itoa(newTunnelName)
	newIfcfgTunnelFileFullPath := tunnelFileMaker.pathToIfcfgTunnelFiles + "ifcfg-" + "tun" + sNewTunnelName

	newRouteTunnelFileFullPath := tunnelFileMaker.pathToIfcfgTunnelFiles + "route-" + sNewTunnelName

	newSysctlConfFileFullPath := tunnelFileMaker.sysctlConfFilePath + sNewTunnelName + "-sysctl.conf"

	enrichedApplicationServer := &domain.ApplicationServer{
		ServerIP:        applicationServers.ServerIP,
		ServerPort:      applicationServers.ServerPort,
		IfcfgTunnelFile: newIfcfgTunnelFileFullPath,
		RouteTunnelFile: newRouteTunnelFileFullPath,
		SysctlConfFile:  newSysctlConfFileFullPath,
		TunnelName:      sNewTunnelName,
	}
	return enrichedApplicationServer, nil
}

// CreateTunnels ...
func (tunnelFileMaker *TunnelFileMaker) CreateTunnels(applicationServers []*domain.ApplicationServer,
	createTunnelUUID string) error {
	for _, applicationServer := range applicationServers {
		if err := tunnelFileMaker.CreateTunnel(applicationServer, createTunnelUUID); err != nil {
			return fmt.Errorf("can't create tunnel: %v", err)
		}
	}
	return nil
}

// CreateTunnel ...
func (tunnelFileMaker *TunnelFileMaker) CreateTunnel(applicationServer *domain.ApplicationServer,
	createTunnelUUID string) error {
	if err := tunnelFileMaker.writeNewTunnelFile(applicationServer, createTunnelUUID); err != nil {
		return fmt.Errorf("can't write new tunnel files: %v", err)
	}

	if err := tunnelFileMaker.ExecuteCommandForTunnel("tun"+applicationServer.TunnelName, "up", createTunnelUUID); err != nil {
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

func (tunnelFileMaker *TunnelFileMaker) writeNewTunnelFile(applicationServer *domain.ApplicationServer,
	createTunnelUUID string) error {
	newDataForTunnelFile := strings.ReplaceAll(rawDataForTunnelFile, "TUNNEL_NAME", "tun"+applicationServer.TunnelName)
	newDataForTunnelFile = strings.ReplaceAll(newDataForTunnelFile, "REAL_SERVER_IP", applicationServer.ServerIP)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("creating new ifcfg for tunnel file: %v", applicationServer.IfcfgTunnelFile)
	err := ioutil.WriteFile(applicationServer.IfcfgTunnelFile, []byte(newDataForTunnelFile+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("can't write new tunnekt to file %v, got error: %v", applicationServer.IfcfgTunnelFile, err)
	}

	dataForTunnelRouteFile := strings.ReplaceAll(rawDataForTunnelRouteFile, "REAL_SERVER_IP", applicationServer.ServerIP)
	dataForTunnelRouteFile = strings.ReplaceAll(dataForTunnelRouteFile, "TUNNEL_NAME", "tun"+applicationServer.TunnelName)

	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("creating new route for tunnel file: %v", applicationServer.RouteTunnelFile)
	err = ioutil.WriteFile(applicationServer.RouteTunnelFile, []byte(dataForTunnelRouteFile+"\n"), 0644) // !!
	if err != nil {
		return fmt.Errorf("can't write new tunnell route to file %v, got error: %v", tunnelFileMaker.pathToIfcfgTunnelFiles+"ifcfg-"+applicationServer.TunnelName, err)
	}

	newRowForSysctlConf := strings.ReplaceAll(rowForSysctlConf, "TUNNEL_NAME", "tun"+applicationServer.TunnelName)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("new sysctl config file name: %v", applicationServer.SysctlConfFile)
	err = ioutil.WriteFile(applicationServer.SysctlConfFile, []byte(newRowForSysctlConf+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("can't write sysctl conf %v, got error: %v", applicationServer.SysctlConfFile, err)
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
func (tunnelFileMaker *TunnelFileMaker) RemoveTunnels(applicationServers []*domain.ApplicationServer,
	removeTunnelUUID string) error {
	for _, applicationServer := range applicationServers {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": removeTunnelUUID,
		}).Debugf("remove tunnel %v:%v files: %v; %v; %v", applicationServer.ServerIP, applicationServer.ServerPort, applicationServer.IfcfgTunnelFile, applicationServer.RouteTunnelFile, applicationServer.SysctlConfFile)
		if err := tunnelFileMaker.RemoveTunnel(applicationServer, removeTunnelUUID); err != nil {
			return fmt.Errorf("can't remove tunnel files: %v", err)
		}
	}
	return nil
}

// RemoveTunnel ...
func (tunnelFileMaker *TunnelFileMaker) RemoveTunnel(applicationServer *domain.ApplicationServer,
	removeTunnelUUID string) error {
	var err error
	if err = tunnelFileMaker.removeFile(applicationServer.IfcfgTunnelFile, removeTunnelUUID); err != nil {
		return err
	}
	if err = tunnelFileMaker.removeFile(applicationServer.RouteTunnelFile, removeTunnelUUID); err != nil {
		return err
	}
	if err = tunnelFileMaker.removeFile(applicationServer.SysctlConfFile, removeTunnelUUID); err != nil {
		return err
	}

	if err := tunnelFileMaker.ExecuteCommandForTunnel("tun"+applicationServer.TunnelName, "down", removeTunnelUUID); err != nil {
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
