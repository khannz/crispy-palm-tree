package domain

import "sync"

// Locker lock other commands for execute
type Locker struct {
	sync.Mutex
}

// GracefullShutdown if need graceful shutdown
type GracefullShutdown struct {
	sync.Mutex
	ShutdownNow  bool
	UsecasesJobs int
}
