package item

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
)

func (s *itemService) DeleteItem(ctx context.Context, userID int64, sku uint32) error {
	if userID <= 0 || sku == 0 {
		return xerrors.ErrInvalidInput
	}
	return s.cartRepository.DeleteItem(ctx, userID, sku)
}
