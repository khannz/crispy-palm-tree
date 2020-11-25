package portadapter

import (
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/executor"
	"github.com/vishvananda/netlink"
)

const routeName = "route worker"

// TODO: much more logs

const ( // TODO: optimize regex
	regexRuleForGetAllLinks     = `tun\d*\b`
	regexRuleForGetRawAllRoutes = `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}(\sdev\stun\d*\s)`
	regexRuleForGetIpRoute      = `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}` // match 1
	regexRuleForGetTableRoute   = `tun(\d*)`                                                                            // match 1 group 1
)

// RouteEntity ...
type RouteEntity struct {
	sync.Mutex
	logging *logrus.Logger
}

// NewRouteEntity ...
func NewRouteEntity(logging *logrus.Logger) *RouteEntity {
	return &RouteEntity{logging: logging}
}

func (routeEntity *RouteEntity) AddRoute(hcDestIP string, hcTunDestIP string, id string) error {
	routeEntity.Lock()
	defer routeEntity.Unlock()

	mask := "/32"
	hcTunDestNetIP, _, err := net.ParseCIDR(hcTunDestIP + mask)
	if err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcDestIP+mask, err)
		return err
	}

	hcDestNetIP, hcDestNetIPNet, err := net.ParseCIDR(hcDestIP + mask)
	if err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcDestIP+mask, err)
		return err
	}

	tun := IP4toInt(hcTunDestNetIP)
	table := int(tun)
	rawTunnelName := strconv.FormatInt(tun, 10)
	tunnelName := "tun" + rawTunnelName
	routeEntity.logging.WithFields(logrus.Fields{
		"entity":   routeName,
		"event id": id,
	}).Tracef("tunnel name: %v", rawTunnelName)

	links, err := getAllLinks()
	if err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("get all links fail: %v", err)
		return err
	}
	if err := routeEntity.addAndUpLink(hcTunDestNetIP, tunnelName, links, id); err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("add and up link fail: %v", err)
		return err
	}
	if err := routeEntity.createRoute(hcDestNetIP, hcDestNetIPNet, rawTunnelName, tunnelName, table, id); err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("create route fail: %v", err)
		return err
	}

	if err := addIPRuleFwmark(table); err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("add ip rule fwmark fail: %v", err)
		return err
	}

	if err := newRpFilter(tunnelName); err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("ne rp filter fail: %v", err)
		return err
	}

	return nil
}

func IP4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
}

func (routeEntity *RouteEntity) addAndUpLink(hcTunDestNetIP net.IP, tunnelName string, links map[string]struct{}, id string) error {
	_, inMap := links[tunnelName]
	if inMap {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
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

func (routeEntity *RouteEntity) createRoute(hcDestNetIP net.IP, hcDestNetIPNet *net.IPNet, rawTunnelName, tunnelName string, table int, id string) error {
	isRouteExist, err := isRouteExist(hcDestNetIP.String(), rawTunnelName)
	if err != nil {
		return err
	} else if isRouteExist {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Tracef("route %v for table %v already exist", hcDestNetIP.String(), rawTunnelName)
		return nil
	}

	linkInfo, err := netlink.LinkByName(tunnelName)
	if err != nil {
		return fmt.Errorf("can't get link by name %v for add route: %v", tunnelName, err)
	}

	route := &netlink.Route{
		LinkIndex: linkInfo.Attrs().Index,
		Dst:       hcDestNetIPNet,
		Table:     table,
	}
	return netlink.RouteAdd(route)
}

func addIPRuleFwmark(tableAndMark int) error {
	family := 2 // ipv4 hardcoded
	rules, err := netlink.RuleList(family)
	if err != nil {
		return fmt.Errorf("can't get current rules: %v", err)
	}
	for _, r := range rules {
		if r.Mark == tableAndMark &&
			r.Table == tableAndMark {
			return nil // rule exist
		}
	}

	rule := netlink.NewRule()
	rule.Mark = tableAndMark
	rule.Table = tableAndMark
	return netlink.RuleAdd(rule)
}

func newRpFilter(tunnelName string) error {
	time.Sleep(time.Duration(100 * time.Millisecond)) // TODO: what until tun up and syscall add that
	args := []string{"-w", "net.ipv4.conf." + tunnelName + ".rp_filter=0"}
	_, _, exitCode, err := executor.Execute("sysctl", "", args)

	if err != nil || exitCode != 0 {
		return fmt.Errorf("error when execute command: sysctl -w net.ipv4.conf.%v.rp_filter=0: %v; exit code: %v",
			tunnelName,
			err,
			exitCode)
	}

	return nil
}

func isRouteExist(hcDestIP string, rawTunnelName string) (bool, error) {
	args := []string{"route", "show", hcDestIP, "table", rawTunnelName}
	stdout, _, exitCode, err := executor.Execute("ip", "", args)
	if err != nil {
		return false, fmt.Errorf("error when execute command: ip route show %v table %v: %v; exit code: %v",
			hcDestIP,
			rawTunnelName,
			err,
			exitCode)
	}

	switch exitCode {
	case 0:
		// TODO: regex check here
		if strings.Contains(string(stdout), hcDestIP) && strings.Contains(string(stdout), rawTunnelName) {
			return true, nil
		}
	case 2:
		return false, nil
	default:
		return false, fmt.Errorf("exit code when execute command: ip route show %v table %v: %v",
			hcDestIP,
			rawTunnelName,
			exitCode)
	}
	return false, nil
}

func (routeEntity *RouteEntity) RemoveRoute(hcDestIP string, hcTunDestIP string, needRemoveTunnel bool, id string) error {
	routeEntity.Lock()
	defer routeEntity.Unlock()

	mask := "/32"
	hcTunDestNetIP, _, err := net.ParseCIDR(hcTunDestIP + mask)
	if err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcDestIP+mask, err)
		return err
	}

	_, hcDestNetIPNet, err := net.ParseCIDR(hcDestIP + mask)
	if err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcDestIP+mask, err)
		return err
	}

	tun := IP4toInt(hcTunDestNetIP)
	table := int(tun)
	rawTunnelName := strconv.FormatInt(tun, 10)
	tunnelName := "tun" + rawTunnelName

	links, err := getAllLinks()
	if err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("get all links fail: %v", err)
		return err
	}
	if err := routeEntity.removeRoute(hcDestNetIPNet, table, rawTunnelName, tunnelName, id); err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("remove route fail: %v", err)
		return err
	}

	if needRemoveTunnel {
		if err := routeEntity.downAndRemoveOldLink(tunnelName, links, id); err != nil {
			routeEntity.logging.WithFields(logrus.Fields{
				"entity":   routeName,
				"event id": id,
			}).Errorf("down and remove link fail: %v", err)
			return err
		}
	}

	if err := delIPRuleFwmark(table); err != nil {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Errorf("remove ip rule fwmark fail: %v", err)
		return err
	}
	return nil
}

func (routeEntity *RouteEntity) downAndRemoveOldLink(tunnelName string, links map[string]struct{}, id string) error {
	_, inMap := links[tunnelName]
	if !inMap {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
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

func (routeEntity *RouteEntity) removeRoute(hcDestNetIPNet *net.IPNet, table int, rawTunnelName, tunnelName string, id string) error {
	isRouteExist, err := isRouteExist(hcDestNetIPNet.IP.String(), rawTunnelName)
	if err != nil {
		return err
	} else if !isRouteExist {
		routeEntity.logging.WithFields(logrus.Fields{
			"entity":   routeName,
			"event id": id,
		}).Tracef("route already not exist")
		return nil
	}

	linkInfo, err := netlink.LinkByName(tunnelName)
	if err != nil {
		return fmt.Errorf("can't get link onfo for remove route for application server %v: %v", hcDestNetIPNet.IP, err)
	}

	route := &netlink.Route{
		LinkIndex: linkInfo.Attrs().Index,
		Dst:       hcDestNetIPNet,
		Table:     table,
	}

	if err := netlink.RouteDel(route); err != nil {
		return err
	}
	return nil
}

func delIPRuleFwmark(tableAndMark int) error {
	family := 2 // ipv4 hardcoded
	rules, err := netlink.RuleList(family)
	if err != nil {
		return fmt.Errorf("can't get current rules: %v", err)
	}
	var ruleExist bool
	for _, r := range rules {
		if r.Mark == tableAndMark &&
			r.Table == tableAndMark {
			ruleExist = true // rule exist
		}
	}

	if ruleExist {
		rule := netlink.NewRule()
		rule.Mark = tableAndMark
		rule.Table = tableAndMark
		return netlink.RuleDel(rule)
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

func (routeEntity *RouteEntity) GetRouteRuntimeConfig(id string) ([]string, error) {
	routeEntity.Lock()
	defer routeEntity.Unlock()
	rawRoutes, err := executeForGetAllRoutes()
	if err != nil {
		return nil, err
	}

	ruleForGetRawAllRoutes := regexp.MustCompile(regexRuleForGetRawAllRoutes)
	rr := ruleForGetRawAllRoutes.FindAllStringSubmatch(rawRoutes, -1)
	rawRoutesMap := make(map[string]struct{}, len(rr))
	for _, r := range rr {
		if len(r) != 0 {
			rawRoutesMap[r[0]] = struct{}{}
		}
	}

	ruleForGetIpRoute := regexp.MustCompile(regexRuleForGetIpRoute)
	ruleForGetTableRoute := regexp.MustCompile(regexRuleForGetTableRoute)
	routesArray := make([]string, 0, len(rawRoutesMap))
	for rawRoute := range rawRoutesMap {
		rawIP := ruleForGetIpRoute.FindAllStringSubmatch(rawRoute, -1)
		rawTunnelName := ruleForGetTableRoute.FindAllStringSubmatch(rawRoute, -1)
		if len(rawIP) != len(rawTunnelName) {
			return nil, fmt.Errorf("some go wrong at regex: ips: %v' tunnel names: %v", len(rawIP), len(rawTunnelName))
		}

		var ip string
		if len(rawIP) != 0 {
			if len(rawIP[0]) != 0 {
				ip = rawIP[0][0]
			}
		}

		var tunnelName string
		if len(rawTunnelName) != 0 {
			if len(rawTunnelName[0]) != 0 {
				tunnelName = rawTunnelName[0][1]
			}
		}
		routesArray = append(routesArray, ip+":"+tunnelName)
	}

	return routesArray, nil
}

func executeForGetAllRoutes() (string, error) {
	args := []string{"route", "list", "table", "all"}
	stdout, _, exitCode, err := executor.Execute("ip", "", args)
	if err != nil || exitCode != 0 {
		return "", fmt.Errorf("error when execute command: route list  table all: %v; exit code: %v",
			err,
			exitCode)
	}

	return string(stdout), nil
}
