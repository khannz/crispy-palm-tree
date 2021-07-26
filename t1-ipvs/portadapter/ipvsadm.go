package portadapter

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tehnerd/gnl2go"
)

const (
	fallbackFlag           uint32 = 8
	fallbackAndshPortFlags uint32 = 24 // Bitwise OR fallbackFlag(8) | shPortFlag(16)
)

type Entity struct {
	sync.Mutex
	logging *logrus.Logger
}

func NewIPVSADMEntity(logging *logrus.Logger) (*IPVSADMEntity, error) {
	ipvs, err := ipvsInit()
	if err != nil {
		return nil, fmt.Errorf("got error when init ipvsadm: %v", err)
	}
	return &IPVSADMEntity{logging: logging}, nil
}

// TODO refactor types?
func (e *Entity) NewIPVSService(
	id, balanceType, vip string,
	protocol, port uint16,
	routingType uint32,
) error {
	isServiceExist, _, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		return err
	}

	if isServiceExist {
		return nil
	}

	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return err
	}
	defer ipvs.Exit()


	err = ipvs.AddServiceWithFlags(vip, port, protocol, newBalanceType, flags)
	if err != nil {
		return err
	}

	return nil
}

func (e *Entity) AddIPVSApplicationServersForService(
	id, balanceType, vip string,
	protocol, port uint16,
	routingType uint32,
	applicationServers map[string]uint16,
) error {

	isServiceExist, pool, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		return err
	}

	if !isServiceExist {
		err := fmt.Errorf("service %v:%v not exist, can't add application servers", vip, port)
		return err
	}

	_, notExistingApplicationServers := e.diffApplicationServersInService(applicationServers, pool) // no lock
	if len(notExistingApplicationServers) == 0 {
		return nil
	}

	for realAddr, realPort := range notExistingApplicationServers {
	}

	if err = e.addApplicationServersToService(vip, port, protocol, routingType, notExistingApplicationServers); err != nil { // lock inside
		return err
	}

	return nil
}

func (e *Entity) RemoveIPVSService(
	id, vip string,
	protocol, port uint16,
) error {
	isServiceExist, _, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		return err
	}
	if !isServiceExist {
		ipvsadmEntity.logging.WithFields(logrus.Fields{
			"entity":   ipvsadmName,
			"event id": id,
		}).Infof("not new servers (at map %v) for add to service %v:%v", applicationServers, vip, port)
		return nil
	}

	ipvsadmEntity.logging.WithFields(logrus.Fields{
		"entity":   ipvsadmName,
		"event id": id,
	}).Infof("new servers (at map %v) at  service %v:%v", notExistingApplicationServers, vip, port)
	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return err
	}
	defer ipvs.Exit()

	if err := ipvs.DelService(vip, port, protocol); err != nil {
		return err
	}

	return nil
}

func (e *Entity) RemoveIPVSApplicationServersFromService(
	id, balanceType, vip string,
	protocol, port uint16,
	routingType uint32,
	applicationServers map[string]uint16,
) error {
	isServiceExist, pool, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		return err
	}

	if !isServiceExist {
		err := fmt.Errorf("service %v:%v not exist, can't add application servers", vip, port)
		return err
	}

	existingApplicationServers, _ := e.diffApplicationServersInService(applicationServers, pool) // no lock
	if len(existingApplicationServers) == 0 {
		return nil
	}

	if err = e.removeIPVSApplicationServersFromService(vip, port, protocol, existingApplicationServers); err != nil { // lock inside
		return err
	}

	return nil
}

func (e *Entity) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	// FIXME I don't get why we need this reading? Result is not even used in any operation
	pools, err := e.getPools() // lock inside
	if err != nil {
		return nil, err
	}

	poolsMap := make(map[string]map[string]uint16, len(pools))
	for _, pool := range pools {
		reals := make(map[string]uint16, len(pool.Dests))
		for _, dest := range pool.Dests {
			reals[dest.IP] = dest.Port
		}
		poolsMap[pool.Service.VIP+":"+strconv.Itoa(int(pool.Service.Port))] = reals
	}

	return poolsMap, nil
}

func (e *Entity) addApplicationServersToService(
	serviceIP string, servicePort uint16, protocol uint16, routingType uint32,
	applicationServers map[string]uint16,
) error {
	// TODO refactor to accept only one real at once
	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return err
	}
	defer ipvs.Exit()

	for realAddr, realPort := range applicationServers {
		err := ipvs.AddDestPort(serviceIP, servicePort, realAddr, realPort, protocol, 10, routingType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Entity) removeIPVSApplicationServersFromService(
	serviceIP string, servicePort uint16, protocol uint16,
	applicationServers map[string]uint16,
) error {
	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return err
	}
	defer ipvs.Exit()

	for realAddr, realPort := range applicationServers {
		err := ipvs.DelDestPort(serviceIP, servicePort, realAddr, realPort, protocol)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Entity) getPools() ([]gnl2go.Pool, error) {
	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return nil, err
	}
	defer ipvs.Exit()

	return ipvs.GetPools()
}

func chooseFlags(balanceType string) (string, []byte) {
	switch balanceType {
	case "mhf":
		return "mh", gnl2go.U32ToBinFlags(fallbackFlag)
	case "mhp":
		return "mh", gnl2go.U32ToBinFlags(fallbackAndshPortFlags)
	default:
		return balanceType, nil
	}
}

func (ipvsadmEntity *IPVSADMEntity) isServiceExist(vip string, port uint16) (bool, gnl2go.Pool, error) {
	var pool gnl2go.Pool
	pools, err := ipvsadmEntity.getPools() // lock inside
	if err != nil {
		return false, pool, fmt.Errorf("can't read actual config: %v", err)
	}

	for _, pool := range pools {
		if pool.Service.VIP == vip &&
			pool.Service.Port == port {
			return true, pool, nil
		}
	}
	return false, pool, nil
}

func (e *Entity) diffApplicationServersInService(reals map[string]uint16, pool gnl2go.Pool) (map[string]uint16, map[string]uint16) {
	existingReals := make(map[string]uint16)
	newReals := make(map[string]uint16)

	for realAddr, realPort := range reals {
		for _, dest := range pool.Dests {
			if realAddr == dest.IP && realPort == dest.Port {
				//glog.Infoln("existing real with ip", ip, "and port", port) // TODO
				existingReals[realAddr] = realPort
			} else {
				//glog.Infoln("new real with ip", ip, "and port", port) // TODO
				newReals[realAddr] = realPort
			}
		}
	}

	return existingReals, newReals
}

// TODO what does it do?
// looks like author does some braindead check... geez
func client(logger *zerolog.Logger) (*gnl2go.IpvsClient, error) {
	logger.Info().Msg("starting ipvs client init...")

	client := new(gnl2go.IpvsClient)
	err := client.Init()
	if err != nil {
		logger.Warn().Msg("ipvs client init failed")
		return nil, err
	}

	//_, err = ipvs.GetPools()
	//if err != nil {
	//	return ipvs, fmt.Errorf("error while running ipvs GetPools method %v", err)
	//}

	logger.Info().Msg("ipvs client init finished")
	return client, nil
}
