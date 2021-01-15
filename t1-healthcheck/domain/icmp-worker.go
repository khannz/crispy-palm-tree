package domain

import (
	"time"
)

// ICMPWorker ...
type ICMPWorker interface {
	IsIcmpCheckOk(string,
		time.Duration,
		int,
		string) bool
}
