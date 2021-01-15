package portadapter

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	icmpwrapper "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/portadapter/icmp-wrapper"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const icmpHealthcheckName = "icmp healthcheck"

// FIXME: SetsockoptTimeval ?

const (
	ICMP4_ECHO_REQUEST = 8
	ICMP4_ECHO_REPLY   = 0
	// ICMP6_ECHO_REQUEST = 128
	ICMP6_ECHO_REPLY = 129
)

type IcmpEntity struct {
	logging *logrus.Logger
	sync.Mutex
	seq int
}

func NewIcmpEntity(logging *logrus.Logger) *IcmpEntity {
	return &IcmpEntity{logging: logging, seq: 0}
}

func (icmpEntity *IcmpEntity) IsIcmpCheckOk(ipS string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	icmpEntity.Lock()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  icmpEntity.seq,
			Data: []byte(""),
		},
	}
	icmpEntity.increaseSeq()
	icmpEntity.Unlock()
	echo, err := m.Marshal(nil)
	if err != nil {
		icmpEntity.logging.WithFields(logrus.Fields{
			"entity":   icmpHealthcheckName,
			"event id": id,
		}).Errorf("icpm for %v Marshal error: %v", ipS, err)
	}

	network := "ip4:icmp"
	ip := net.ParseIP(ipS)
	if ip == nil {
		icmpEntity.logging.WithFields(logrus.Fields{
			"entity":   icmpHealthcheckName,
			"event id": id,
		}).Errorf("icpm invalid address %v: %v", ipS, err)
	}
	err = exchangeICMPEcho(network, ip, timeout, echo, fwmark)
	if err != nil {
		icmpEntity.logging.WithFields(logrus.Fields{
			"entity":   icmpHealthcheckName,
			"event id": id,
		}).Tracef("icmp exchangeICMPEchoerror: %v", err)
		return false
	}

	return true
}

func exchangeICMPEcho(network string, ip net.IP, timeout time.Duration, echo []byte, fwmark int) error {
	c, err := icmpwrapper.ListenPacket(network, "", fwmark)
	if err != nil {
		return fmt.Errorf("faled ListenPacket: %v", err)
	}
	defer c.Close() // FIXME: check sockets not running out

	_, err = c.WriteTo(echo, &net.UDPAddr{IP: ip})
	if err != nil {
		return fmt.Errorf("faled icpm wrapper WriteTo: %v", err)
	}

	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("can't set deadline: %v", err)
	}
	reply := make([]byte, 256)

	for {
		n, _, err := c.ReadFrom(reply)
		if err != nil {
			return err
		}

		rm, err := icmp.ParseMessage(1, reply[:n])
		if err != nil {
			return err
		}
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			return nil
		default:
			return fmt.Errorf("unsoported reply type: %v", rm.Type)
		}
	}
}

func (icmpEntity *IcmpEntity) increaseSeq() {
	// icmpEntity must be locked
	if icmpEntity.seq == 65535 {
		icmpEntity.seq = 1
	} else {
		icmpEntity.seq++
	}
}
