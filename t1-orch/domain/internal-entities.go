package domain

import "sync"

// Locker lock other commands for execute
type Locker struct {
	sync.Mutex
}

// GracefulShutdown if need graceful shutdown
type GracefulShutdown struct {
	sync.Mutex
	ShutdownNow  bool
	UsecasesJobs int
}
