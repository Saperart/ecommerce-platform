package cart

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
)

func (s *cartService) CheckoutCart(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, xerrors.ErrInvalidInput
	}
	items, err := s.cartRepository.GetItemsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, xerrors.ErrCartIsEmpty
	}
	orderID, err := s.lomsClient.CreateOrder(ctx, userID, items)
	if err != nil {
		return 0, err
	}
	if err := s.cartRepository.DeleteItemsByUserID(ctx, userID); err != nil {
		return 0, err
	}
	return orderID, nil
}
