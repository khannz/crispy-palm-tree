package portadapter

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/executor"
	"github.com/vishvananda/netlink"
)

// TODO: logic refactor. files != tunnel. also commands for routing != create tunnels

const tunnelFileMakerEntityName = "tunnel-file-maker"

const rowForSysctlConf = `net.ipv4.conf.TUNNEL_NAME.rp_filter=0`

// TunnelFileMaker ...
type TunnelFileMaker struct {
	sysctlConfFilePath string
	isMockMode         bool
	logging            *logrus.Logger
}

// NewTunnelFileMaker ...
func NewTunnelFileMaker(sysctlConfFilePath string,
	isMockMode bool,
	logging *logrus.Logger) *TunnelFileMaker {
	return &TunnelFileMaker{
		sysctlConfFilePath: sysctlConfFilePath,
		isMockMode:         isMockMode,
		logging:            logging,
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
	newSysctlConfFileFullPath := tunnelFileMaker.sysctlConfFilePath + sNewTunnelName + "-sysctl.conf"

	tunnelFilesInfo.TunnelName = sNewTunnelName
	tunnelFilesInfo.SysctlConfFile = newSysctlConfFileFullPath

	if err := tunnelFileMaker.writeNewTunnelFile(tunnelFilesInfo, // TODO: remove that
		createTunnelUUID); err != nil {
		return fmt.Errorf("can't write new tunnel files: %v", err)
	}

	table := 10   // TODO: remove hardcode
	mask := "/32" // TODO: remove hardcode

	if err := tunnelFileMaker.addAndUpNewLink("tun"+sNewTunnelName, tunnelFilesInfo.ApplicationServerIP+"/32"); err != nil {
		return fmt.Errorf("can't up tunnel: %v", err)
	}

	if err := tunnelFileMaker.addRoute(sNewTunnelName, tunnelFilesInfo.ApplicationServerIP, mask, table); err != nil {
		return fmt.Errorf("can't create route: %v", err)
	}

	return nil
}

func (tunnelFileMaker *TunnelFileMaker) chooseNewTunnelName() (int, error) {
	var nextTunnelName int

	files, err := ioutil.ReadDir(tunnelFileMaker.sysctlConfFilePath)
	if err != nil {
		return nextTunnelName, fmt.Errorf("read dir %v, got error %v", tunnelFileMaker.sysctlConfFilePath, err)
	}
	if len(files) == 0 {
		return nextTunnelName, nil
	}

	var sliceOfOldTunelNames []int
	for _, f := range files {
		tunnelFileMaker.logging.Warnf("file find: %v", f.Name())
		if strings.Contains(f.Name(), "-sysctl.conf") {
			stringOldTunelName := strings.TrimSuffix(f.Name(), "-sysctl.conf")
			intOldTunelName, err := strconv.Atoi(stringOldTunelName)
			if err != nil {
				tunnelFileMaker.logging.Warnf("invalid sysctl.conf '%v' name in %v", tunnelFileMaker.sysctlConfFilePath, f.Name())
				continue
			}
			sliceOfOldTunelNames = append(sliceOfOldTunelNames, intOldTunelName)
		}
	}

	if len(sliceOfOldTunelNames) > 0 { // TODO: take last "free"
		sort.Sort(sort.Reverse(sort.IntSlice(sliceOfOldTunelNames)))
		nextTunnelName = sliceOfOldTunelNames[0] + 1
	}
	return nextTunnelName, nil
}

func (tunnelFileMaker *TunnelFileMaker) writeNewTunnelFile(tunnelFilesInfo *domain.TunnelForApplicationServer,
	createTunnelUUID string) error {
	newRowForSysctlConf := strings.ReplaceAll(rowForSysctlConf,
		"TUNNEL_NAME",
		"tun"+tunnelFilesInfo.TunnelName)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":     tunnelFileMakerEntityName,
		"event uuid": createTunnelUUID,
	}).Tracef("new sysctl config file name: %v", tunnelFilesInfo.SysctlConfFile)
	if err := ioutil.WriteFile(tunnelFilesInfo.SysctlConfFile, []byte(newRowForSysctlConf+"\n"), 0644); err != nil {
		return fmt.Errorf("can't write sysctl conf %v, got error: %v",
			tunnelFilesInfo.SysctlConfFile, err)
	}
	return nil
}

func (tunnelFileMaker *TunnelFileMaker) addAndUpNewLink(tunnelName, applicationServerIPAndMask string) error {
	ipNew, _, err := net.ParseCIDR(applicationServerIPAndMask)
	if err != nil {
		return fmt.Errorf("parse ip from %v fail: %v", applicationServerIPAndMask, err)
	}

	linkNew := &netlink.Iptun{LinkAttrs: netlink.LinkAttrs{Name: tunnelName}, Remote: ipNew}
	err = netlink.LinkAdd(linkNew)
	if err != nil {
		tunnelFileMaker.logging.Warnf("can't LinkAdd for tunnel %v: %v", tunnelName, err)
		// return fmt.Errorf("can't LinkAdd for tunnel %v: %v", tunnelName, err)
	}
	if err := netlink.LinkSetUp(linkNew); err != nil {
		return fmt.Errorf("can't LinkSetUp for device %v: %v", tunnelName, err)
	}
	return nil
}

// TODO: global refactor func argument call (like remove route and add, give ip+mask or only ip)
func (tunnelFileMaker *TunnelFileMaker) addRoute(sNewTunnelName, applicationServerIP, mask string, table int) error {
	linkInfo, err := netlink.LinkByName("tun" + sNewTunnelName)
	if err != nil {
		return fmt.Errorf("can't get link onfo for add route for application server %v: %v", applicationServerIP, err)
	}

	_, destination, err := net.ParseCIDR(applicationServerIP + mask)
	if err != nil {
		return fmt.Errorf("parse ip from %v fail: %v", applicationServerIP+mask, err)
	}
	route := &netlink.Route{
		LinkIndex: linkInfo.Attrs().Index,
		Dst:       destination,
		Table:     table,
	}
	return netlink.RouteAdd(route)
}

// ip link delete dev tun100 && rm -f /etc/sysctl.d/100-sysctl.conf
// cat /sys/devices/virtual/net/tun100/ifindex
func (tunnelFileMaker *TunnelFileMaker) downAndRemoveOldLink(tunnelName string) error {
	linkOld, err := netlink.LinkByName(tunnelName)
	if err != nil {
		return fmt.Errorf("can't get LinkByName %v: %v", tunnelName, err)
	}
	if err := netlink.LinkSetDown(linkOld); err != nil {
		return fmt.Errorf("can't LinkSetDown for device %v: %v", tunnelName, err)
	}

	err = netlink.LinkDel(linkOld)
	if err != nil {
		return fmt.Errorf("can't LinkDel for device %v: %v", tunnelName, err)
	}
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
		}).Debugf("remove tunnel %v; file: %v", tunnelFilesInfo.ApplicationServerIP, tunnelFilesInfo.SysctlConfFile)

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

	table := 10
	mask := "/32" // TODO: remove hardcode
	if err := tunnelFileMaker.removeRoute(tunnelFilesInfo.ApplicationServerIP, mask, table, tunnelFilesInfo.TunnelName); err != nil {
		return fmt.Errorf("can't remove route: %v", err) // FIXME: here broken. maybe removed some already?
	}

	if err := tunnelFileMaker.downAndRemoveOldLink("tun" + tunnelFilesInfo.TunnelName); err != nil {
		return fmt.Errorf("can't remove tunnel: %v", err)
	}

	if err = tunnelFileMaker.removeFile(tunnelFilesInfo.SysctlConfFile, removeTunnelUUID); err != nil {
		return err
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

// TODO: fix broken netlink lib
func (tunnelFileMaker *TunnelFileMaker) removeRoute(applicationServerIP, mask string, table int, tunnelName string) error {
	fullCommand := "ip"
	contextFolder := ""
	arguments := []string{"link", "del", "tun" + tunnelName}
	stdout, stderr, exitCode, err := executor.Execute(fullCommand, contextFolder, arguments)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("error code %v. stdout: %v; stderr: %v", exitCode, string(stdout), string(stderr))
	}

	// linkInfo, err := netlink.LinkByName("tun" + tunnelName)
	// if err != nil {
	// 	return fmt.Errorf("can't get link onfo for add route for application server %v: %v", applicationServerIP, err)
	// }

	// _, destination, err := net.ParseCIDR(applicationServerIP + mask)
	// if err != nil {
	// 	return fmt.Errorf("parse ip from %v fail: %v", applicationServerIP+mask, err)
	// }
	// route := &netlink.Route{
	// 	LinkIndex: linkInfo.Attrs().Index,
	// 	Dst:       destination,
	// 	Table:     table,
	// }
	// if err := netlink.RouteDel(route); err != nil {
	// 	return fmt.Errorf("netlink can't delete route %v: %v", route, err)
	// }

	return nil
}
