package domain

import (
	"time"
)

// HTTPAndHTTPSWorker ...
type HTTPAndHTTPSWorker interface {
	IsHttpOrHttpsCheckOk(string,
		string,
		map[int]struct{},
		time.Duration,
		int,
		bool,
		string) bool
}
