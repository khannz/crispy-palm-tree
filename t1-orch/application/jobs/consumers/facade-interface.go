package consumers

import (
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
)

//FacadeInterface facade interface
type FacadeInterface interface {
	ApplyNewConfig(updatedServicesInfo domain.ServiceInfoConf) error
	RemoveAllConfigs() error
}
