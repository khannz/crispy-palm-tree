package portadapter

import (
	"context"
	"net"
	"time"
)

func makeDialer(addr string, t time.Duration) func(ctx context.Context, addr string) (net.Conn, error) {
	f := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial("unix", addr)
	}

	return func(ctx context.Context, addr string) (net.Conn, error) {
		if deadline, ok := ctx.Deadline(); ok {
			return f(addr, time.Until(deadline))
		}
		return f(addr, 0)
	}
}
