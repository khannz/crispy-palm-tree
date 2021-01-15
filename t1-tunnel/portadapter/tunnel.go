package portadapter

import (
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/executor"
	"github.com/vishvananda/netlink"
)

const tunnelWorkerName = "tunnel worker"

// TODO: much more logs

const (
	regexRuleForGetAllLinks = `tun\d*\b`
)

// TunnelEntity ...
type TunnelEntity struct {
	sync.Mutex
	logging *logrus.Logger
}

// NewTunnelEntity ...
func NewTunnelEntity(logging *logrus.Logger) *TunnelEntity {
	return &TunnelEntity{logging: logging}
}

func (tunnelEntity *TunnelEntity) AddTunnel(hcTunDestIP string, id string) error {
	tunnelEntity.Lock()
	defer tunnelEntity.Unlock()

	mask := "/32"
	hcTunDestNetIP, _, err := net.ParseCIDR(hcTunDestIP + mask)
	if err != nil {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcTunDestIP+mask, err)
		return err
	}

	tun := IP4toInt(hcTunDestNetIP)

	rawTunnelName := strconv.FormatInt(tun, 10)
	tunnelName := "tun" + rawTunnelName
	tunnelEntity.logging.WithFields(logrus.Fields{
		"entity":   tunnelWorkerName,
		"event id": id,
	}).Tracef("tunnel name: %v", rawTunnelName)

	links, err := getAllLinks()
	if err != nil {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("get all links fail: %v", err)
		return err
	}
	if err := tunnelEntity.addAndUpLink(hcTunDestNetIP, tunnelName, links, id); err != nil {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("add and up link fail: %v", err)
		return err
	}

	go newRpFilter(tunnelName, id, tunnelEntity.logging) // TODO: rework reverse path filtering: 1) syscall 2) no sleeps

	return nil
}

func IP4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
}

func (tunnelEntity *TunnelEntity) addAndUpLink(hcTunDestNetIP net.IP, tunnelName string, links map[string]struct{}, id string) error {
	_, inMap := links[tunnelName]
	if inMap {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Tracef("tunnel name %v already in links: %v", tunnelName, links)
		return nil
	}

	linkNew := &netlink.Iptun{LinkAttrs: netlink.LinkAttrs{Name: tunnelName},
		Remote: hcTunDestNetIP}
	if err := netlink.LinkAdd(linkNew); err != nil {
		return fmt.Errorf("can't LinkAdd for tunnel %v: %v", tunnelName, err)
	}
	if err := netlink.LinkSetUp(linkNew); err != nil {
		return fmt.Errorf("can't LinkSetUp for device %v: %v", tunnelName, err)
	}
	return nil
}

func newRpFilter(tunnelName string, id string, logging *logrus.Logger) {
	time.Sleep(time.Duration(100 * time.Millisecond)) // TODO: what until tun up and syscall add that
	args := []string{"-w", "net.ipv4.conf." + tunnelName + ".rp_filter=0"}
	_, _, exitCode, err := executor.Execute("sysctl", "", args)

	if err != nil || exitCode != 0 {
		logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("set rp filter fail: error when execute command: sysctl -w net.ipv4.conf.%v.rp_filter=0: %v; exit code: %v",
			tunnelName,
			err,
			exitCode)
	}
}

func (tunnelEntity *TunnelEntity) RemoveTunnel(hcTunDestIP string, id string) error {
	tunnelEntity.Lock()
	defer tunnelEntity.Unlock()

	mask := "/32"
	hcTunDestNetIP, _, err := net.ParseCIDR(hcTunDestIP + mask)
	if err != nil {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcTunDestIP+mask, err)
		return err
	}

	tun := IP4toInt(hcTunDestNetIP)
	rawTunnelName := strconv.FormatInt(tun, 10)
	tunnelName := "tun" + rawTunnelName

	links, err := getAllLinks()
	if err != nil {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("get all links fail: %v", err)
		return err
	}

	if err := tunnelEntity.downAndRemoveOldLink(tunnelName, links, id); err != nil {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Errorf("down and remove link fail: %v", err)
		return err
	}

	return nil
}

func (tunnelEntity *TunnelEntity) downAndRemoveOldLink(tunnelName string, links map[string]struct{}, id string) error {
	_, inMap := links[tunnelName]
	if !inMap {
		tunnelEntity.logging.WithFields(logrus.Fields{
			"entity":   tunnelWorkerName,
			"event id": id,
		}).Tracef("tunnel name %v not in links: %v", tunnelName, links)
		return nil
	}

	linkOld, err := netlink.LinkByName(tunnelName)
	if err != nil {
		return nil
	}
	if err := netlink.LinkSetDown(linkOld); err != nil {
		return nil
	}

	err = netlink.LinkDel(linkOld)
	if err != nil {
		return fmt.Errorf("can't LinkDel for device %v: %v", tunnelName, err)
	}
	return nil
}

func getAllLinks() (map[string]struct{}, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	linksNames := make(map[string]struct{})
	tunRule := regexp.MustCompile(regexRuleForGetAllLinks)
	for _, link := range linkList {
		if tunRule.MatchString(link.Attrs().Name) {
			linksNames[link.Attrs().Name] = struct{}{}
		}
	}

	return linksNames, nil
}

func (tunnelEntity *TunnelEntity) GetTunnelRuntime(id string) (map[string]struct{}, error) {
	tunnelEntity.Lock()
	defer tunnelEntity.Unlock()
	return getAllLinks()
}
