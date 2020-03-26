package portadapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

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
	logging                *logrus.Logger
}

// NewTunnelFileMaker ...
func NewTunnelFileMaker(pathToIfcfgTunnelFiles string,
	sysctlConfFilePath string,
	logging *logrus.Logger) *TunnelFileMaker {
	return &TunnelFileMaker{
		pathToIfcfgTunnelFiles: pathToIfcfgTunnelFiles,
		sysctlConfFilePath:     sysctlConfFilePath,
		logging:                logging,
	}
}

// CreateTunnel ...
func (tunnelFileMaker *TunnelFileMaker) CreateTunnel(deployedEntities map[string][]string,
	realServersData map[string]string,
	newNWBRequestUUID string) (map[string][]string, error) {
	nextTunnelName, err := tunnelFileMaker.chooseNewTunnelName()
	if err != nil {
		return nil, fmt.Errorf("can't choose new tunnel name: %v", err)
	}

	newTunnels, createdTunnelFiles, err := tunnelFileMaker.writeNewTunnelFiles(realServersData, nextTunnelName, newNWBRequestUUID)
	if err != nil {
		return deployedEntities, fmt.Errorf("can't write new tunnel files: %v", err)
	}
	deployedEntities["createdTunnelFiles"] = createdTunnelFiles

	err = tunnelFileMaker.ExecuteCommandForTunnels(newTunnels, "up", newNWBRequestUUID)
	if err != nil {
		return deployedEntities, fmt.Errorf("can't execute command for up tunnel: %v", err)
	}
	deployedEntities["newTunnels"] = newTunnels

	return deployedEntities, nil
}

func (tunnelFileMaker *TunnelFileMaker) chooseNewTunnelName() (int, error) {
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

func (tunnelFileMaker *TunnelFileMaker) writeNewTunnelFiles(realServersData map[string]string,
	nextTunnelName int,
	newNWBRequestUUID string) ([]string,
	[]string,
	error) {
	newTunnels := []string{}
	createdTunnelFiles := []string{}
	for realServerIP := range realServersData {
		var newDataForTunnelFile string
		newTunnelName := "tun" + strconv.Itoa(nextTunnelName)
		newTunnels = append(newTunnels, newTunnelName)
		newDataForTunnelFile = strings.ReplaceAll(rawDataForTunnelFile, "TUNNEL_NAME", newTunnelName)
		newDataForTunnelFile = strings.ReplaceAll(newDataForTunnelFile, "REAL_SERVER_IP", realServerIP)

		newIfcfgTunnelFileFullPath := tunnelFileMaker.pathToIfcfgTunnelFiles + "ifcfg-" + newTunnelName
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": newNWBRequestUUID,
		}).Tracef("creating new ifcfg for tunnel file: %v", newIfcfgTunnelFileFullPath)
		err := ioutil.WriteFile(newIfcfgTunnelFileFullPath, []byte(newDataForTunnelFile+"\n"), 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("can't write new tunnekt to file %v, got error: %v", newIfcfgTunnelFileFullPath, err)
		}
		createdTunnelFiles = append(createdTunnelFiles, newIfcfgTunnelFileFullPath)

		dataForTunnelRouteFile := strings.ReplaceAll(rawDataForTunnelRouteFile, "REAL_SERVER_IP", realServerIP)
		dataForTunnelRouteFile = strings.ReplaceAll(dataForTunnelRouteFile, "TUNNEL_NAME", newTunnelName)

		newRouteTunnelFileFullPath := tunnelFileMaker.pathToIfcfgTunnelFiles + "route-" + newTunnelName
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": newNWBRequestUUID,
		}).Tracef("creating new route for tunnel file: %v", newRouteTunnelFileFullPath)
		err = ioutil.WriteFile(newRouteTunnelFileFullPath, []byte(dataForTunnelRouteFile+"\n"), 0644) // !!
		if err != nil {
			return nil, nil, fmt.Errorf("can't write new tunnell route to file %v, got error: %v", tunnelFileMaker.pathToIfcfgTunnelFiles+"ifcfg-"+newTunnelName, err)
		}
		createdTunnelFiles = append(createdTunnelFiles, newRouteTunnelFileFullPath)

		newRowForSysctlConf := strings.ReplaceAll(rowForSysctlConf, "TUNNEL_NAME", newTunnelName)
		newSysctlConfFileName := strconv.Itoa(nextTunnelName) + "-sysctl.conf"

		newSysctlConfFileFullPath := tunnelFileMaker.sysctlConfFilePath + newSysctlConfFileName
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": newNWBRequestUUID,
		}).Tracef("new sysctl config file name: %v", newSysctlConfFileFullPath)
		err = ioutil.WriteFile(newSysctlConfFileFullPath, []byte(newRowForSysctlConf+"\n"), 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("can't write sysctl conf %v, got error: %v", newSysctlConfFileFullPath, err)
		}
		createdTunnelFiles = append(createdTunnelFiles, newSysctlConfFileFullPath)
		nextTunnelName++
	}
	return newTunnels, createdTunnelFiles, nil
}

// ExecuteCommandForTunnels ...
func (tunnelFileMaker *TunnelFileMaker) ExecuteCommandForTunnels(newTunnels []string, arg string, newNWBRequestUUID string) error {
	for _, newTunnel := range newTunnels {
		// tunnelFileMaker.logging.WithFields(logrus.Fields{
		// 	"entity":     tunnelFileMakerEntityName,
		// 	"event uuid": newNWBRequestUUID,
		// }).Debugf("mock of execute ip command for tunnel %v", newTunnel)
		args := []string{"link", "set", "dev", newTunnel, arg}
		stdout, stderr, exitCode, err := executor.Execute("/usr/sbin/ip", "", args)
		if err != nil {
			return fmt.Errorf("when execute command %v, got error: %v", "/usr/sbin/ip link set dev "+newTunnel+" "+arg, err)
		}
		if exitCode != 0 {
			return fmt.Errorf("when execute command %v, got exit code != 0: stdout: %v, stderr: %v, exitCode: %v",
				"/usr/sbin/ip link set dev "+newTunnel+" "+arg,
				string(stdout),
				string(stderr),
				string(exitCode))
		}
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":     tunnelFileMakerEntityName,
			"event uuid": newNWBRequestUUID,
		}).Debugf("result of execute ip command: stdout: %v, stderr: %v, exitCode: %v", string(stdout), string(stderr), string(exitCode))
	}
	return nil
}

// RemoveCreatedTunnelFiles ...
func (tunnelFileMaker *TunnelFileMaker) RemoveCreatedTunnelFiles(createdTunnelFiles []string, requestUUID string) error {
	for _, tunnelFile := range createdTunnelFiles {
		err := os.Remove(tunnelFile)
		if err != nil {
			tunnelFileMaker.logging.WithFields(logrus.Fields{
				"entity":     tunnelFileMakerEntityName,
				"event uuid": requestUUID,
			}).Errorf("can't remove tunnel file %v, got error: %v", tunnelFile, err)
		}
	}
	return nil
}

// func DetectAllTunnels

// DetectTunnels ... TODO: detect only first "lucky" for real server, that is bad
func (tunnelFileMaker *TunnelFileMaker) DetectTunnels(realServersData map[string]string,
	deployedEntities map[string][]string,
	removeRequestUUID string) (map[string][]string, error) {
	allFilesInTunnelDir, err := ioutil.ReadDir(tunnelFileMaker.pathToIfcfgTunnelFiles)
	if err != nil {
		return deployedEntities, fmt.Errorf("can't read dir %v, got error %v",
			tunnelFileMaker.pathToIfcfgTunnelFiles,
			err)
	}

	allTunnelRouteFiles := getAllTunnelRouteFiles(allFilesInTunnelDir, tunnelFileMaker.pathToIfcfgTunnelFiles)
	realServersDataIPAddresses := getRealServersDataIPAddresses(realServersData)

	tunnelFiles, err := tunnelFileMaker.getAllTunnelsFilesPath(realServersDataIPAddresses, allTunnelRouteFiles)
	if err != nil {
		return deployedEntities, fmt.Errorf("error when get all tunnels files path: %v", err)
	}
	deployedEntities["createdTunnelFiles"] = tunnelFiles

	return deployedEntities, nil
}

func getRealServersDataIPAddresses(realServersData map[string]string) []string {
	realServersDataIPAddresses := []string{}
	for ip := range realServersData {
		realServersDataIPAddresses = append(realServersDataIPAddresses, ip)
	}
	return realServersDataIPAddresses
}

func getAllTunnelRouteFiles(allFilesInTunnelDir []os.FileInfo, pathToDir string) []string {
	allTunnelRouteFiles := []string{}
	for _, file := range allFilesInTunnelDir {
		if strings.Contains(file.Name(), "route-tun") {
			allTunnelRouteFiles = append(allTunnelRouteFiles, pathToDir+file.Name())
		}
	}
	return allTunnelRouteFiles
}

func (tunnelFileMaker *TunnelFileMaker) getAllTunnelsFilesPath(realServersDataIPAddresses, allTunnelRouteFiles []string) ([]string, error) {
	routeFilesForRemove, err := filesContains(realServersDataIPAddresses, allTunnelRouteFiles)
	if err != nil {
		return nil, fmt.Errorf("got error when search ip tunnels %v in files %v: %v",
			realServersDataIPAddresses,
			allTunnelRouteFiles,
			err)
	}

	tunnelFilesForRemove := getTunnelsFileForRemove(routeFilesForRemove)

	sysctlConfFilesForRemove, err := tunnelFileMaker.getSysctlConfFilesForRemove(routeFilesForRemove)
	if err != nil {
		return nil, fmt.Errorf("can't get sysctl conf files for remove: %v", err)
	}

	return combineThreeSlices(routeFilesForRemove, tunnelFilesForRemove, sysctlConfFilesForRemove), nil
}

func getTunnelsFileForRemove(routeFilesForRemove []string) []string {
	tunnelFilesForRemove := []string{}
	for _, findedRouteFileForRemove := range routeFilesForRemove {
		tunnelFileForRemove := strings.ReplaceAll(findedRouteFileForRemove, "route-", "ifcfg-")
		tunnelFilesForRemove = append(tunnelFilesForRemove, tunnelFileForRemove)
	}
	return tunnelFilesForRemove
}

func (tunnelFileMaker *TunnelFileMaker) getSysctlConfFilesForRemove(routeFilesForRemove []string) ([]string, error) {
	sysctlConfFilesForRemove := []string{}
	for _, findedRouteFileForRemove := range routeFilesForRemove {
		re := regexp.MustCompile(`route-tun(.*)`)
		finded := re.FindAllStringSubmatch(findedRouteFileForRemove, -1)
		if len(finded) >= 1 {
			if len(finded[0]) >= 2 {
				tunnelNumber := finded[0][1]
				sysctlConfFileForRemove := tunnelFileMaker.sysctlConfFilePath + tunnelNumber + "-sysctl.conf"
				sysctlConfFilesForRemove = append(sysctlConfFilesForRemove, sysctlConfFileForRemove)
			} else {
				return sysctlConfFilesForRemove, fmt.Errorf("can't find tunnel number in filename %v", findedRouteFileForRemove)
			}
		} else {
			return sysctlConfFilesForRemove, fmt.Errorf("can't find tunnel number in filename %v", findedRouteFileForRemove)
		}
	}
	return sysctlConfFilesForRemove, nil
}

func combineThreeSlices(sliceOne, sliceTwo, sliceThree []string) []string {
	resultSlice := append(sliceOne, sliceTwo...)
	return append(resultSlice, sliceThree...)
}
