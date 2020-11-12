package healthchecks

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

type conn struct {
	fd      int
	f       *os.File
	netConn net.Conn
}

func dialTCP(network, ipS string, port string, timeout time.Duration, mark int) (nc net.Conn, err error) {
	ip := net.ParseIP(ipS)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address %q", ip)
	}
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port number %q", port)
	}

	var domain int
	var rsa syscall.Sockaddr

	domain = syscall.AF_INET
	if ip.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 address %q", ip)
	}
	sa := &syscall.SockaddrInet4{Port: int(p)}
	copy(sa.Addr[:], ip.To4())
	rsa = sa

	c := &conn{}

	defer func() {
		if err != nil {
			if c.netConn != nil { // awoid segmentation violation (do not close nil)
				c.netConn.Close()
			}
		}
	}()

	c.fd, err = syscall.Socket(domain, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}

	if mark != 0 {
		if err := setSocketMark(c.fd, mark); err != nil {
			return nil, err
		}
	}

	if err := setSocketTimeout(c.fd, timeout); err != nil {
		return nil, err
	}
	for {
		err := syscall.Connect(c.fd, rsa)
		if err == nil {
			break
		}
		// Blocking socket connect may be interrupted with EINTR
		if err != syscall.EINTR {
			return nil, os.NewSyscallError("connect", err)
		}
	}
	if err := setSocketTimeout(c.fd, 0); err != nil {
		return nil, err
	}

	lsa, _ := syscall.Getsockname(c.fd)
	rsa, _ = syscall.Getpeername(c.fd)
	name := fmt.Sprintf("%s %s -> %s", network, sockaddrToString(lsa), sockaddrToString(rsa))
	c.f = os.NewFile(uintptr(c.fd), name)

	// When we call net.FileConn the socket will be made non-blocking and
	// we will get a *net.TCPConn in return. The *os.File needs to be
	// closed in addition to the *net.TCPConn when we're done (conn.Close
	// takes care of that for us).
	if c.netConn, err = net.FileConn(c.f); err != nil {
		return nil, err
	}
	if _, ok := c.netConn.(*net.TCPConn); !ok {
		return nil, fmt.Errorf("%T is not a *net.TCPConn", c.netConn)
	}

	return c.netConn, nil
}

// setSocketMark sets packet marking on the given socket.
func setSocketMark(fd, mark int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_MARK, mark); err != nil {
		return os.NewSyscallError("failed to set mark", err)
	}
	return nil
}

// setSocketTimeout sets the receive and send timeouts on the given socket.
func setSocketTimeout(fd int, timeout time.Duration) error {
	tv := syscall.NsecToTimeval(timeout.Nanoseconds())
	for _, opt := range []int{syscall.SO_RCVTIMEO, syscall.SO_SNDTIMEO} {
		if err := syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, opt, &tv); err != nil {
			return os.NewSyscallError("setsockopt", err)
		}
	}
	return nil
}

func sockaddrToString(sa syscall.Sockaddr) string {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	case *syscall.SockaddrInet6:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	default:
		return fmt.Sprintf("(unknown - %T)", sa)
	}
}
