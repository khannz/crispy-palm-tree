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

	if err := tunnelFileMaker.writeNewTunnelFile(tunnelFilesInfo,
		createTunnelUUID); err != nil {
		return fmt.Errorf("can't write new tunnel files: %v", err)
	}

	table := 10   // TODO: remove hardcode
	mask := "/32" // TODO: remove hardcode
	if err := tunnelFileMaker.addRoute(newTunnelName, tunnelFilesInfo.ApplicationServerIP, mask, table); err != nil {
		return fmt.Errorf("can't create route: %v", err)
	}

	if err := tunnelFileMaker.addAndUpNewLink("tun"+sNewTunnelName, tunnelFilesInfo.ApplicationServerIP+"/32"); err != nil {
		return fmt.Errorf("can't up tunnel: %v", err)
	}

	return nil
}

func (tunnelFileMaker *TunnelFileMaker) chooseNewTunnelName() (int, error) { // TODO: rework that, it's too hard just for new number
	files, err := ioutil.ReadDir(tunnelFileMaker.sysctlConfFilePath)
	if err != nil {
		return 0, fmt.Errorf("read dir %v, got error %v", tunnelFileMaker.sysctlConfFilePath, err)
	}
	if len(files) == 0 {
		return 0, nil
	}

	var sliceOfOldTunelNames []string
	for _, f := range files {
		if strings.Contains(f.Name(), "-sysctl.conf") {
			sliceOfOldTunelNames = append(sliceOfOldTunelNames, strings.TrimPrefix(f.Name(), "-sysctl.conf"))
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
		return fmt.Errorf("can't LinkAdd for tunnel %v: %v", tunnelName, err)
	}
	if err := netlink.LinkSetUp(linkNew); err != nil {
		return fmt.Errorf("can't LinkSetUp for device %v: %v", tunnelName, err)
	}
	return nil
}

func (tunnelFileMaker *TunnelFileMaker) addRoute(newTunnelName int, applicationServerIP, mask string, table int) error {
	_, destination, err := net.ParseCIDR(applicationServerIP + mask)
	if err != nil {
		return fmt.Errorf("parse ip from %v fail: %v", applicationServerIP+mask, err)
	}
	route := &netlink.Route{
		LinkIndex: newTunnelName,
		Dst:       destination,
		Table:     table,
	}
	return netlink.RouteAdd(route)
}

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

	if err = tunnelFileMaker.removeFile(tunnelFilesInfo.SysctlConfFile, removeTunnelUUID); err != nil {
		return err
	}

	if err := tunnelFileMaker.downAndRemoveOldLink("tun" + tunnelFilesInfo.TunnelName); err != nil {
		return fmt.Errorf("can't remove tunnel: %v", err)
	}

	mask := "/32" // TODO: remove hardcode
	if err := tunnelFileMaker.removeRoute(tunnelFilesInfo.ApplicationServerIP, mask); err != nil {
		return fmt.Errorf("can't remove route: %v", err)
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

func (tunnelFileMaker *TunnelFileMaker) removeRoute(applicationServerIP, mask string) error {
	destination, _, err := net.ParseCIDR(applicationServerIP + mask)
	if err != nil {
		return fmt.Errorf("parse ip from %v fail: %v", applicationServerIP+mask, err)
	}
	routes, err := netlink.RouteGet(destination)
	if err != nil {
		return fmt.Errorf("netlink can't get routes by %v", applicationServerIP+mask)
	}
	for _, route := range routes {
		if err := netlink.RouteDel(&route); err != nil {
			return fmt.Errorf("netlink can't delete route %v: %v", route, err)
		}
	}
	return nil
}
