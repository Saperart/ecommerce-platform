package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (s *lomsService) CreateOrder(
	ctx context.Context,
	userID int64,
	items []entity.OrderItem,
) (int64, error) {
	if userID <= 0 || len(items) == 0 {
		return 0, xerrors.ErrInvalidInput
	}
	for _, item := range items {
		if item.SKU == 0 || item.Count == 0 {
			return 0, xerrors.ErrInvalidInput
		}
	}

	var orderID int64
	err := s.withTx(ctx, func(ctx context.Context) error {
		if err := s.stocksRepository.ReserveStocks(ctx, items); err != nil {
			return err
		}

		var err error
		orderID, err = s.orderRepository.CreateOrder(ctx, userID, items)
		if err != nil {
			return err
		}

		return s.enqueueOrderStatusChanged(ctx, userID, orderID, entity.OrderStatusAwaitingPayment)
	})
	if err != nil {
		return 0, err
	}

	return orderID, nil
}
