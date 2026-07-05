package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (s *lomsService) GetOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
	if orderID <= 0 {
		return nil, xerrors.ErrInvalidInput
	}
	order, err := s.orderRepository.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}
