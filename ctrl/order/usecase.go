package order

import (
	"context"
	"github.com/khannz/crispy-palm-tree/ctrl/models"
	"time"
)

type UseCase interface {
	CreateOrder(ctx context.Context, order *models.Order, oid, src, raw, smid string, typ int, createdat time.Time) error
	GetOrders(ctx context.Context, order *models.Order) ([]*models.Order, error)
}
