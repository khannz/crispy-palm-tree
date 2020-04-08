package domain

import "sync"

// Locker lock other commands for execute
type Locker struct {
	sync.Mutex
}
