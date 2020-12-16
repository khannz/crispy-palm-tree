package portadapter

import (
	"fmt"
	"math/big"
	"net"
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/executor"
	"github.com/vishvananda/netlink"
)

const ipRuleName = "ipRule worker"

// TODO: much more logs

const ( // TODO: optimize regex
	regexRuleForGetAllLinks      = `tun\d*\b`
	regexRuleForGetRawAllIPRules = `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}(\sdev\stun\d*\s)`
	regexRuleForGetIpIPRule      = `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}` // match 1
	regexRuleForGetTableIPRule   = `tun(\d*)`                                                                            // match 1 group 1
)

// IPRuleEntity ...
type IPRuleEntity struct {
	sync.Mutex
	logging *logrus.Logger
}

// NewIPRuleEntity ...
func NewIPRuleEntity(logging *logrus.Logger) *IPRuleEntity {
	return &IPRuleEntity{logging: logging}
}

func (ipRuleEntity *IPRuleEntity) AddIPRule(hcTunDestIP string, id string) error {
	ipRuleEntity.Lock()
	defer ipRuleEntity.Unlock()

	mask := "/32"
	hcTunDestNetIP, _, err := net.ParseCIDR(hcTunDestIP + mask)
	if err != nil {
		ipRuleEntity.logging.WithFields(logrus.Fields{
			"entity":   ipRuleName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcTunDestIP+mask, err)
		return err
	}

	tun := IP4toInt(hcTunDestNetIP)
	table := int(tun)

	if err := addIPRuleFwmark(table); err != nil {
		ipRuleEntity.logging.WithFields(logrus.Fields{
			"entity":   ipRuleName,
			"event id": id,
		}).Errorf("add ip rule fwmark fail: %v", err)
		return err
	}

	return nil
}

func IP4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
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

func (ipRuleEntity *IPRuleEntity) RemoveIPRule(hcTunDestIP string, id string) error {
	ipRuleEntity.Lock()
	defer ipRuleEntity.Unlock()

	mask := "/32"
	hcTunDestNetIP, _, err := net.ParseCIDR(hcTunDestIP + mask)
	if err != nil {
		ipRuleEntity.logging.WithFields(logrus.Fields{
			"entity":   ipRuleName,
			"event id": id,
		}).Errorf("parse ip from %v fail: %v", hcTunDestIP+mask, err)
		return err
	}

	tun := IP4toInt(hcTunDestNetIP)
	table := int(tun)

	if err := delIPRuleFwmark(table); err != nil {
		ipRuleEntity.logging.WithFields(logrus.Fields{
			"entity":   ipRuleName,
			"event id": id,
		}).Errorf("remove ip rule fwmark fail: %v", err)
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

func (ipRuleEntity *IPRuleEntity) GetIPRuleRuntimeConfig(id string) (map[string]struct{}, error) {
	ipRuleEntity.Lock()
	defer ipRuleEntity.Unlock()
	rawIPRules, err := executeForGetAllIPRules()
	if err != nil {
		return nil, err
	}

	ruleForGetRawAllIPRules := regexp.MustCompile(regexRuleForGetRawAllIPRules)
	rr := ruleForGetRawAllIPRules.FindAllStringSubmatch(rawIPRules, -1)
	rawIPRulesMap := make(map[string]struct{}, len(rr))
	for _, r := range rr {
		if len(r) != 0 {
			rawIPRulesMap[r[0]] = struct{}{}
		}
	}

	fmt.Println(rawIPRulesMap)

	return nil, nil
}

func executeForGetAllIPRules() (string, error) {
	args := []string{"ipRule", "list", "table", "all"}
	stdout, _, exitCode, err := executor.Execute("ip", "", args)
	if err != nil || exitCode != 0 {
		return "", fmt.Errorf("error when execute command: ipRule list  table all: %v; exit code: %v",
			err,
			exitCode)
	}

	return string(stdout), nil
}
