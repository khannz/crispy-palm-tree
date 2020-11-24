package domain

import (
	"time"
)

// ICMPWorker ...
type ICMPWorker interface {
	IsIcmpCheckOk(string,
		int,
		time.Duration,
		int,
		string) bool
}
