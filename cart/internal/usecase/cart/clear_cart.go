package cart

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
)

func (s *cartService) ClearCart(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return xerrors.ErrInvalidInput
	}

	return s.cartRepository.DeleteItemsByUserID(ctx, userID)
}
