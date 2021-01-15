package icmpwrapper

import (
	"net"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// From https://godoc.org/golang.org/x/net/internal/iana, can't import internal packages
const (
	ianaProtocolICMP     = 1
	ianaProtocolIPv6ICMP = 58
)

func ListenPacket(network, address string, mark int) (*PacketConn, error) {
	var family, proto int
	switch network {
	case "udp4":
		family, proto = syscall.AF_INET, ianaProtocolICMP
	case "udp6":
		family, proto = syscall.AF_INET6, ianaProtocolIPv6ICMP
	default:
		i := last(network, ':')
		if i < 0 {
			i = len(network)
		}

		switch network[:i] {
		case "ip4":
			proto = ianaProtocolICMP
		case "ip6":
			proto = ianaProtocolIPv6ICMP
		}

		switch network[i+1:] {
		case "icmp":
			family = 2
		case "icmp6":
			family = 10
		default:
			return nil, net.InvalidAddrError("unexpected family for icmp (supported: icmp, icmp6)")
		}
	}
	var cerr error
	var c net.PacketConn

	switch family {
	case syscall.AF_INET, syscall.AF_INET6:
		s, err := syscall.Socket(family, syscall.SOCK_DGRAM, syscall.IPPROTO_ICMP)
		if err != nil {
			return nil, os.NewSyscallError("socket", err)
		}
		if mark != 0 {
			if err := setSocketMark(s, mark); err != nil {
				return nil, err
			}
		}

		sa, err := sockaddr(family, address)
		if err != nil {
			syscall.Close(s)
			return nil, err
		}
		if err := syscall.Bind(s, sa); err != nil {
			syscall.Close(s)
			return nil, os.NewSyscallError("bind", err)
		}
		f := os.NewFile(uintptr(s), "datagram-oriented icmp")
		c, cerr = net.FilePacketConn(f)
		f.Close()
	default:
		c, cerr = net.ListenPacket(network, address)
	}
	if cerr != nil {
		return nil, cerr
	}

	switch proto {
	case ianaProtocolICMP:
		return &PacketConn{c: c, p4: ipv4.NewPacketConn(c)}, nil
	case ianaProtocolIPv6ICMP:
		return &PacketConn{c: c, p6: ipv6.NewPacketConn(c)}, nil
	default:
		return &PacketConn{c: c}, nil
	}
}

func sockaddr(family int, address string) (syscall.Sockaddr, error) {
	switch family {
	case syscall.AF_INET:
		a, err := net.ResolveIPAddr("ip4", address)
		if err != nil {
			return nil, err
		}
		if len(a.IP) == 0 {
			a.IP = net.IPv4zero
		}
		if a.IP = a.IP.To4(); a.IP == nil {
			return nil, net.InvalidAddrError("non-ipv4 address")
		}
		sa := &syscall.SockaddrInet4{}
		copy(sa.Addr[:], a.IP)
		return sa, nil
	case syscall.AF_INET6:
		a, err := net.ResolveIPAddr("ip6", address)
		if err != nil {
			return nil, err
		}
		if len(a.IP) == 0 {
			a.IP = net.IPv6unspecified
		}
		if a.IP.Equal(net.IPv4zero) {
			a.IP = net.IPv6unspecified
		}
		if a.IP = a.IP.To16(); a.IP == nil || a.IP.To4() != nil {
			return nil, net.InvalidAddrError("non-ipv6 address")
		}
		sa := &syscall.SockaddrInet6{ZoneId: zoneToUint32(a.Zone)}
		copy(sa.Addr[:], a.IP)
		return sa, nil
	default:
		return nil, net.InvalidAddrError("unexpected family")
	}
}

func zoneToUint32(zone string) uint32 {
	if zone == "" {
		return 0
	}
	if ifi, err := net.InterfaceByName(zone); err == nil {
		return uint32(ifi.Index)
	}
	n, err := strconv.Atoi(zone)
	if err != nil {
		return 0
	}
	return uint32(n)
}

func last(s string, b byte) int {
	i := len(s)
	for i--; i >= 0; i-- {
		if s[i] == b {
			break
		}
	}
	return i
}

// setSocketMark sets packet marking on the given socket.
func setSocketMark(fd, mark int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_MARK, mark); err != nil {
		return os.NewSyscallError("failed to set mark", err)
	}
	return nil
}
