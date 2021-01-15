package domain

import (
	"time"
)

// HTTPAndHTTPSWorker ...
type HTTPAndHTTPSWorker interface {
	IsHttpOrHttpsCheckOk(string,
		time.Duration,
		int,
		bool,
		string) bool
}
