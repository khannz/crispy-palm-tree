package portadapter

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// DummyEntity ...
type DummyEntity struct {
	sync.Mutex
	interfaceName string
	logging       *logrus.Logger
}

// NewDummyEntity ...
func NewDummyEntity(interfaceName string, logging *logrus.Logger) *DummyEntity {
	return &DummyEntity{interfaceName: interfaceName, logging: logging}
}

func (dummyEntity *DummyEntity) AddToDummy(serviceIP string) error {
	dummyEntity.Lock()
	defer dummyEntity.Unlock()

	addrs, err := dummyEntity.getDummyAddrs()
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
		if err := dummyEntity.addAddr(incomeIPAndMask); err != nil {
			return fmt.Errorf("can't add ip addr %v, got err %v", incomeIPAndMask, err)
		}
	}
	return nil
}
func (dummyEntity *DummyEntity) RemoveFromDummy(serviceIP string) error {
	dummyEntity.Lock()
	defer dummyEntity.Unlock()

	addrs, err := dummyEntity.getDummyAddrs()
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
		if err := dummyEntity.removeAddr(incomeIPAndMask); err != nil {
			return fmt.Errorf("can't remove ip addr %v, got err %v", incomeIPAndMask, err)
		}
	}
	return nil
}

func (dummyEntity *DummyEntity) getDummyAddrs() ([]net.Addr, error) {
	i, err := net.InterfaceByName(dummyEntity.interfaceName)
	if err != nil {
		return nil, fmt.Errorf("can't get InterfaceByName: %v", err)
	}
	addrs, err := i.Addrs()
	if err != nil {
		return nil, fmt.Errorf("can't addrs in interfaces: %v", err)
	}
	return addrs, err
}

func (dummyEntity *DummyEntity) removeAddr(addrForDel string) error {
	dummy, err := netlink.LinkByName(dummyEntity.interfaceName)
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

func (dummyEntity *DummyEntity) addAddr(addrForAdd string) error {
	dummy, err := netlink.LinkByName(dummyEntity.interfaceName)
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
