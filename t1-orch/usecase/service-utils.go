package usecase

import (
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

func decreaseJobs(gracefulShutdown *domain.GracefulShutdown) {
	gracefulShutdown.Lock()
	defer gracefulShutdown.Unlock()
	gracefulShutdown.UsecasesJobs--
}
