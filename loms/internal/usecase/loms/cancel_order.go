package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (s *lomsService) CancelOrder(ctx context.Context, orderID int64) error {
	if orderID <= 0 {
		return xerrors.ErrInvalidInput
	}

	order, err := s.orderRepository.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if order.Status != entity.OrderStatusAwaitingPayment {
		return xerrors.ErrInvalidOrderStatus
	}

	return s.withTx(ctx, func(ctx context.Context) error {
		if err := s.orderRepository.SetOrderStatus(ctx, orderID, entity.OrderStatusCancelled); err != nil {
			return err
		}
		for _, item := range order.Items {
			if err := s.stocksRepository.ReleaseStock(ctx, item.SKU, uint64(item.Count)); err != nil {
				return err
			}
		}
		return s.enqueueOrderStatusChanged(ctx, order.UserID, orderID, entity.OrderStatusCancelled)
	})
}
