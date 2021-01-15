package domain

import (
	"time"
)

// TCPWorker ...
type TCPWorker interface {
	IsTcpCheckOk(string,
		time.Duration,
		int,
		string) bool
}
