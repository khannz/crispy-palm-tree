package domain

import (
	"time"
)

// HTTPSWorker ...
type HTTPSWorker interface {
	IsHttpsCheckOk(string,
		time.Duration,
		int,
		string) bool
}
