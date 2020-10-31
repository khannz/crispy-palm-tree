package portadapter

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
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
	createTunnelID string) ([]*domain.TunnelForApplicationServer, error) {
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":   tunnelFileMakerEntityName,
		"event id": createTunnelID,
	}).Tracef("starting create tunnels: %v", tunnelsFilesInfo)
	newTunnelsFilesInfo := []*domain.TunnelForApplicationServer{}
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		if tunnelFilesInfo.ServicesToTunnelCount == 0 {
			if err := tunnelFileMaker.CreateTunnel(tunnelFilesInfo, createTunnelID); err != nil {
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
	createTunnelID string) error {
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":   tunnelFileMakerEntityName,
		"event id": createTunnelID,
	}).Tracef("starting create tunnel: %v", *tunnelFilesInfo)
	newTunnelName, err := tunnelFileMaker.chooseNewTunnelName()
	if err != nil {
		return fmt.Errorf("can't choose new tunnel name: %v", err)
	}
	sNewTunnelName := strconv.Itoa(newTunnelName)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":   tunnelFileMakerEntityName,
		"event id": createTunnelID,
	}).Tracef("new tunnel name: %v", sNewTunnelName)
	newSysctlConfFileFullPath := tunnelFileMaker.sysctlConfFilePath + sNewTunnelName + "-sysctl.conf"

	tunnelFilesInfo.TunnelName = sNewTunnelName
	tunnelFilesInfo.SysctlConfFile = newSysctlConfFileFullPath

	if err := tunnelFileMaker.writeNewTunnelFile(tunnelFilesInfo, // TODO: remove that
		createTunnelID); err != nil {
		return fmt.Errorf("can't write new tunnel files: %v", err)
	}

	table := 10   // TODO: remove hardcode
	mask := "/32" // TODO: remove hardcode

	if err := tunnelFileMaker.addAndUpNewLink("tun"+sNewTunnelName, tunnelFilesInfo.ApplicationServerIP+"/32"); err != nil {
		return fmt.Errorf("can't up tunnel: %v", err)
	}

	if err := tunnelFileMaker.addRoute(sNewTunnelName, tunnelFilesInfo.ApplicationServerIP, mask, table, createTunnelID); err != nil {
		return fmt.Errorf("can't create route: %v", err)
	}

	return nil
}

func (tunnelFileMaker *TunnelFileMaker) chooseNewTunnelName() (int, error) {
	nextTunnelName := 100

	files, err := ioutil.ReadDir(tunnelFileMaker.sysctlConfFilePath)
	if err != nil {
		return nextTunnelName, fmt.Errorf("read dir %v, got error %v", tunnelFileMaker.sysctlConfFilePath, err)
	}
	if len(files) == 0 {
		return nextTunnelName, nil
	}

	var sliceOfOldTunelNames []int
	for _, f := range files {
		tunnelFileMaker.logging.Tracef("file find: %v", f.Name())
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
	createTunnelID string) error {
	newRowForSysctlConf := strings.ReplaceAll(rowForSysctlConf,
		"TUNNEL_NAME",
		"tun"+tunnelFilesInfo.TunnelName)
	tunnelFileMaker.logging.WithFields(logrus.Fields{
		"entity":   tunnelFileMakerEntityName,
		"event id": createTunnelID,
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
func (tunnelFileMaker *TunnelFileMaker) addRoute(sNewTunnelName, applicationServerIP, mask string, table int, id string) error {

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

	if err = netlink.RouteAdd(route); err != nil {
		// FIXME: maybe broken. temp dirty fix
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":   tunnelFileMakerEntityName,
			"event id": id,
		}).Warnf("can't add route for tunnel %v; server %v, mask %v, table %v", sNewTunnelName, applicationServerIP, mask, table)
	}
	return nil
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
	removeTunnelID string) ([]*domain.TunnelForApplicationServer, error) {
	newTunnelsFilesInfo := []*domain.TunnelForApplicationServer{}
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":   tunnelFileMakerEntityName,
			"event id": removeTunnelID,
		}).Debugf("remove tunnel %v; file: %v", tunnelFilesInfo.ApplicationServerIP, tunnelFilesInfo.SysctlConfFile)

		if tunnelFilesInfo.ServicesToTunnelCount == 1 {
			if err := tunnelFileMaker.RemoveTunnel(tunnelFilesInfo, removeTunnelID); err != nil {
				return nil, fmt.Errorf("can't remove tunnel files: %v", err)
			}
		}
		tunnelFilesInfo.ServicesToTunnelCount = tunnelFilesInfo.ServicesToTunnelCount - 1
		newTunnelFilesInfo := *tunnelFilesInfo
		newTunnelsFilesInfo = append(newTunnelsFilesInfo, &newTunnelFilesInfo)
	}
	return newTunnelsFilesInfo, nil
}

// RemoveAllTunnels remove tunnels whithout any checks
func (tunnelFileMaker *TunnelFileMaker) RemoveAllTunnels(tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	removeTunnelID string) error {
	var errors []error
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":   tunnelFileMakerEntityName,
			"event id": removeTunnelID,
		}).Debugf("remove tunnel %v; file: %v", tunnelFilesInfo.ApplicationServerIP, tunnelFilesInfo.SysctlConfFile)

		if err := tunnelFileMaker.RemoveTunnel(tunnelFilesInfo, removeTunnelID); err != nil {
			errors = append(errors, err)
		}
	}
	return combineErrors(errors)
}

// RemoveTunnel ...
func (tunnelFileMaker *TunnelFileMaker) RemoveTunnel(tunnelFilesInfo *domain.TunnelForApplicationServer,
	removeTunnelID string) error {
	var err error

	table := 10
	mask := "/32" // TODO: remove hardcode
	if err := tunnelFileMaker.removeRoute(tunnelFilesInfo.ApplicationServerIP, mask, table, tunnelFilesInfo.TunnelName, removeTunnelID); err != nil {
		tunnelFileMaker.logging.Errorf("can't remove route: %v", err)
	}

	if err := tunnelFileMaker.downAndRemoveOldLink("tun" + tunnelFilesInfo.TunnelName); err != nil {
		tunnelFileMaker.logging.Errorf("can't remove tunnel: %v", err)
	}

	if err = tunnelFileMaker.removeFile(tunnelFilesInfo.SysctlConfFile, removeTunnelID); err != nil {
		tunnelFileMaker.logging.Errorf("can't remove route: %v", err)
	}

	return nil
}

func (tunnelFileMaker *TunnelFileMaker) removeFile(filePath, requestID string) error {
	if err := os.Remove(filePath); err != nil {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":   tunnelFileMakerEntityName,
			"event id": requestID,
		}).Errorf("can't remove tunnel file %v, got error: %v", filePath, err)
		return err
	}
	return nil
}

func (tunnelFileMaker *TunnelFileMaker) removeRoute(applicationServerIP, mask string, table int, tunnelName string, id string) error {
	linkInfo, err := netlink.LinkByName("tun" + tunnelName)
	if err != nil {
		return fmt.Errorf("can't get link onfo for remove route for application server %v:%v", applicationServerIP, err)
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
	if err := netlink.RouteDel(route); err != nil {
		tunnelFileMaker.logging.WithFields(logrus.Fields{
			"entity":   tunnelFileMakerEntityName,
			"event id": id,
		}).Warnf("can't remove route for tunnel %v; server %v, mask %v, table %v", tunnelName, applicationServerIP, mask, table)
	}
	return nil
}

func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	var errorsStringSlice []string
	for _, err := range errors {
		errorsStringSlice = append(errorsStringSlice, err.Error())
	}
	return fmt.Errorf(strings.Join(errorsStringSlice, "\n"))
}
