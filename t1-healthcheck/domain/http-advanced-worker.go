package domain

import (
	"time"
)

// HTTPAdvancedWorker ...
type HTTPAdvancedWorker interface {
	IsHttpAdvancedCheckOk(string,
		string,
		bool,
		map[string]string,
		time.Duration,
		int,
		string) bool
}
