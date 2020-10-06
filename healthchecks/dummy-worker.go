package healthchecks

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// TODO: move HC to other service

func (hc *HeathcheckEntity) removeFromDummyWrapper(serviceIP string) {
	hc.dw.Lock()
	defer hc.dw.Unlock()
	if !hc.isMockMode {
		if err := RemoveFromDummy(serviceIP); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": healthcheckID,
			}).Errorf("Heathcheck error: can't remove service ip from dummy: %v", err)
		}
	}
}

func (hc *HeathcheckEntity) addToDummyWrapper(serviceIP string) {
	hc.dw.Lock()
	defer hc.dw.Unlock()
	if !hc.isMockMode {
		if err := addToDummy(serviceIP); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": healthcheckID,
			}).Errorf("Heathcheck error: can't add service ip to dummy: %v", err)
		}
	}
}

func addToDummy(serviceIP string) error {
	addrs, err := getDummyAddrs()
	if err != nil {
		return err
	}
	var addrIsFounded bool
	incomeIPAndMask := serviceIP + "/32"
	for _, addr := range addrs {
		if incomeIPAndMask == addr.String() {
			addrIsFounded = true
			break

		}
	}
	if !addrIsFounded {
		if err := addAddr(incomeIPAndMask); err != nil {
			return fmt.Errorf("can't add ip addr %v, got err %v", incomeIPAndMask, err)
		}
	}
	return nil
}

// RemoveFromDummy remove service from dummy
func RemoveFromDummy(serviceIP string) error {
	addrs, err := getDummyAddrs()
	if err != nil {
		return err
	}
	var addrIsFounded bool
	incomeIPAndMask := serviceIP + "/32"
	for _, addr := range addrs {
		if incomeIPAndMask == addr.String() {
			addrIsFounded = true
			break
		}
	}
	if addrIsFounded {
		if err := removeAddr(incomeIPAndMask); err != nil {
			return fmt.Errorf("can't remove ip addr %v, got err %v", incomeIPAndMask, err)
		}
	}
	return nil
}

func getDummyAddrs() ([]net.Addr, error) {
	i, err := net.InterfaceByName("dummy0")
	if err != nil {
		return nil, fmt.Errorf("can't get InterfaceByName: %v", err)
	}
	addrs, err := i.Addrs()
	if err != nil {
		return nil, fmt.Errorf("can't addrs in interfaces: %v", err)
	}
	return addrs, err
}

func removeAddr(addrForDel string) error {
	dummy, err := netlink.LinkByName("dummy0") // hardcoded
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(addrForDel)
	if err != nil {
		return err
	}

	if err = netlink.AddrDel(dummy, addr); err != nil {
		return err
	}
	return nil
}

func addAddr(addrForAdd string) error {
	dummy, err := netlink.LinkByName("dummy0") // hardcoded
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(addrForAdd)
	if err != nil {
		return err
	}

	if err = netlink.AddrAdd(dummy, addr); err != nil {
		return err
	}
	return nil
}
