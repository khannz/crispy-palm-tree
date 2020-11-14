package order

import (
	"context"
	"github.com/khannz/crispy-palm-tree/ctrl/models"
)

type Order interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrders(ctx context.Context, order *models.Order) ([]*models.Order, error)
}