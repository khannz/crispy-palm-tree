package portadapter

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/rs/zerolog"
	"github.com/tehnerd/gnl2go"
)

const (
	fallbackFlag           uint32 = 8
	fallbackAndshPortFlags uint32 = 24 // Bitwise OR fallbackFlag(8) | shPortFlag(16)
)

type Entity struct {
	sync.Mutex
	logger *zerolog.Logger
}

func NewEntity(logger *zerolog.Logger) *Entity {
	return &Entity{
		logger: logger,
	}
}

// TODO refactor types?
func (e *Entity) NewIPVSService(
	id, balanceType, vip string,
	protocol, port uint16,
	routingType uint32,
) error {
	e.logger.Trace().Msg("starting ipvs service creation...")
	isServiceExist, _, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("existence check was failed")
		return err
	}

	if isServiceExist {
		e.logger.Info().
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("ipvsadm service already exists")
		return nil
	}

	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return err
	}
	defer ipvs.Exit()

	newBalanceType, flags := chooseFlags(balanceType, e.logger)

	err = ipvs.AddServiceWithFlags(vip, port, protocol, newBalanceType, flags)
	if err != nil {
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Str("balance_type", newBalanceType).
			Bytes("flags", flags).
			Msg("can't add IPv4 service")
		return err
	}

	e.logger.Info().
		Uint16("protocol", protocol).
		Str("vip", vip).
		Uint16("port", port).
		Str("balanceType", newBalanceType).
		Msg("vip created")

	return nil
}

func (e *Entity) RemoveIPVSService(
	id, vip string,
	protocol, port uint16,
) error {
	isServiceExist, _, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("existence check for service failed")
		return err
	}
	if !isServiceExist {
		e.logger.Debug().
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("deletion request ineffective since service does not exist")
		return nil
	}

	e.Lock()
	defer e.Unlock()

	ipvs, err := client(e.logger)
	if err != nil {
		return err
	}
	defer ipvs.Exit()

	if err := ipvs.DelService(vip, port, protocol); err != nil {
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("service deletion returned error")
		return err
	}

	e.logger.Info().
		Uint16("protocol", protocol).
		Str("vip", vip).
		Uint16("port", port).
		Msg("vip removed")

	return nil
}

func (e *Entity) AddIPVSApplicationServersForService(
	id, balanceType, vip string,
	protocol, port uint16,
	routingType uint32,
	applicationServers map[string]uint16,
) error {
	e.logger.Info().
		Msg("starting to add reals into ipvs...")

	// FIXME current tuple for isServiceExist doesn't include protocol
	isServiceExist, pool, err := e.isServiceExist(vip, port, protocol) // lock inside
	if err != nil {
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("existence check for vip+port returned error")
		return err
	}

	if !isServiceExist {
		err := fmt.Errorf("service %v:%v not exist, can't add application servers", vip, port)
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("can't add reals for non-existing service")
		return err
	}

	_, notExistingApplicationServers := e.diffApplicationServersInService(applicationServers, pool) // no lock
	if len(notExistingApplicationServers) == 0 {
		e.logger.Trace().
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("no new reals for vip")
		return nil
	}

	for realAddr, realPort := range notExistingApplicationServers {
		e.logger.Trace().
			Str("ip", realAddr).
			Uint16("port", realPort).
			Msgf("got new real for vip %d_%s:%d", protocol, vip, port)
	}

	if err = e.addApplicationServersToService(vip, port, protocol, routingType, notExistingApplicationServers); err != nil { // lock inside
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("can't add reals for vip")
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
		e.logger.Error().
			Err(err).
			Uint16("protocol", protocol).
			Str("vip", vip).
			Uint16("port", port).
			Msg("existence check for vip+port+protocol returned error")
		return err
	}

	if !isServiceExist {
		err := fmt.Errorf("service %v:%v not exist, can't add application servers", vip, port)
		e.logger.Error().
			Err(err).
			Str("vip", vip).
			Uint16("port", port).
			Uint16("protocol", protocol).
			Msg("can't delete reals for non-existing service")
		return err
	}

	existingApplicationServers, _ := e.diffApplicationServersInService(applicationServers, pool) // no lock
	if len(existingApplicationServers) == 0 {
		return nil
	}

	if err = e.removeIPVSApplicationServersFromService(vip, port, protocol, existingApplicationServers); err != nil { // lock inside
		e.logger.Error().
			Err(err).
			Str("vip", vip).
			Uint16("port", port).
			Uint16("protocol", protocol).
			Msg("can't delete reals for vip")
		return err
	}

	return nil
}

func (e *Entity) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	// FIXME I don't get why we need this reading? Result is not even used in any operation
	pools, err := e.getPools() // lock inside
	if err != nil {
		e.logger.Error().
			Err(err).
			Msg("can't read current ipvs state")
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
			e.logger.Error().
				Err(err).
				Uint16("protocol", protocol).
				Str("vip", serviceIP).
				Uint16("port", servicePort).
				Msgf("can't add real %s:%d", realAddr, realPort)
			return err
		}
		e.logger.Info().
			Uint16("protocol", protocol).
			Str("vip", serviceIP).
			Uint16("port", servicePort).
			Msgf("added real %s:%d", realAddr, realPort)
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
			e.logger.Error().
				Err(err).
				Uint16("protocol", protocol).
				Str("vip", serviceIP).
				Uint16("port", servicePort).
				Msgf("can't remove real %s:%d", realAddr, realPort)
			return err
		}
		e.logger.Info().
			Uint16("protocol", protocol).
			Str("vip", serviceIP).
			Uint16("port", servicePort).
			Msgf("removed real %s:%d", realAddr, realPort)
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

func chooseFlags(balanceType string, logger *zerolog.Logger) (string, []byte) {
	switch balanceType {
	case "mhf":
		// TODO
		logger.Debug().
			Bytes("current_flag", gnl2go.U32ToBinFlags(fallbackFlag)).
			Msg("flag values")
		return "mh", gnl2go.U32ToBinFlags(fallbackFlag)
	case "mhp":
		// TODO
		logger.Debug().
			Bytes("current_flag", gnl2go.U32ToBinFlags(fallbackAndshPortFlags)).
			Msg("flag values")
		return "mh", gnl2go.U32ToBinFlags(fallbackAndshPortFlags)
	default:
		return balanceType, nil
	}
}

// FIXME not working after tuple fix
func (e *Entity) isServiceExist(vip string, port, proto uint16) (bool, gnl2go.Pool, error) {
	// FIXME why does pool returning from purely bool func???
	var pool gnl2go.Pool
	pools, err := e.getPools() // lock inside
	if err != nil {
		e.logger.Error().
			Err(err).
			Msg("can't read current ipvs state")
		return false, pool, err
	}

	e.logger.Debug().Msgf("all pools: %v", pools)
	e.logger.Debug().
		Msgf("looking for: %d_%s_%d", proto, vip, port)
	for _, pool := range pools {
		if pool.Service.VIP == vip && pool.Service.Port == port && pool.Service.Proto == proto {
			e.logger.Debug().
				Msgf("got match: %d_%s_%d", pool.Service.Proto, pool.Service.VIP, pool.Service.Port)
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
	c := new(gnl2go.IpvsClient)
	err := c.Init()
	if err != nil {
		logger.Warn().Msg("ipvs client init failed")
		return nil, err
	}

	//_, err = ipvs.GetPools()
	//if err != nil {
	//	return ipvs, fmt.Errorf("error while running ipvs GetPools method %v", err)
	//}
	return c, nil
}
