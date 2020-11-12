package healthchecks

import (
	"fmt"
	"net"
	"os"
	"time"

	icmpwrapper "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/portadapter/healthchecks/icmp-wrapper"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const icmpHealthcheckName = "icmp healthcheck"

const (
	ICMP4_ECHO_REQUEST = 8
	ICMP4_ECHO_REPLY   = 0
	// ICMP6_ECHO_REQUEST = 128
	ICMP6_ECHO_REPLY = 129
)

func IcmpCheckOk(ipS string,
	seq int,
	timeout time.Duration,
	fwmark int,
	id string,
	logging *logrus.Logger) bool {
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte(""),
		},
	}
	echo, err := m.Marshal(nil)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   icmpHealthcheckName,
			"event id": id,
		}).Errorf("icpm for %v Marshal error: %v", ipS, err)
	}

	network := "ip4:icmp"
	ip := net.ParseIP(ipS)
	if ip == nil {
		logging.WithFields(logrus.Fields{
			"entity":   icmpHealthcheckName,
			"event id": id,
		}).Errorf("icpm invalid address %v: %v", ipS, err)
	}
	err = exchangeICMPEcho(network, ip, timeout, echo, fwmark)
	if err != nil {
		logging.WithFields(logrus.Fields{
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
	defer c.Close()

	_, err = c.WriteTo(echo, &net.UDPAddr{IP: ip})
	if err != nil {
		return fmt.Errorf("faled icpm wrapper WriteTo: %v", err)
	}

	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("can't set deadline: %v", err)
	}
	reply := make([]byte, 256)
	for {
		_, addr, err := c.ReadFrom(reply)
		if err != nil {
			return err
		}
		if !ip.Equal(net.ParseIP(addr.String())) {
			continue
		}
		if reply[0] != ICMP4_ECHO_REPLY && reply[0] != ICMP6_ECHO_REPLY {
			continue
		}
		xid, xseqnum, _ := parseICMPEchoReply(echo)
		rid, rseqnum, rchksum := parseICMPEchoReply(reply)
		if rid != xid || rseqnum != xseqnum {
			continue
		}
		if reply[0] == ICMP4_ECHO_REPLY {
			cs := icmpChecksum(reply)
			if cs != 0 {
				return fmt.Errorf("Bad ICMP checksum: %x", rchksum)
			}
		}
		break
	}
	return nil
}

func parseICMPEchoReply(msg []byte) (id, seqnum, chksum uint16) {
	id = uint16(msg[4])<<8 | uint16(msg[5])
	seqnum = uint16(msg[6])<<8 | uint16(msg[7])
	chksum = uint16(msg[2])<<8 | uint16(msg[3])
	return
}

func icmpChecksum(msg []byte) uint16 {
	cklen := len(msg)
	s := uint32(0)
	for i := 0; i < cklen-1; i += 2 {
		s += uint32(msg[i+1])<<8 | uint32(msg[i])
	}
	if cklen&1 == 1 {
		s += uint32(msg[cklen-1])
	}
	s = (s >> 16) + (s & 0xffff)
	s = s + (s >> 16)
	return uint16(^s)
}
