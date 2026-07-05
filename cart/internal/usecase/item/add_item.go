package item

import (
	"context"
	"errors"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
)

func (s *itemService) AddItem(ctx context.Context, userID int64, sku, count uint32) error {
	if userID <= 0 || sku == 0 || count == 0 {
		return xerrors.ErrInvalidInput
	}

	if _, err := s.productClient.GetProduct(ctx, sku); err != nil {
		if errors.Is(err, xerrors.ErrProductNotFound) {
			return fmt.Errorf("get product info error: %w", xerrors.ErrProductNotFound)
		}
		return err
	}

	available, err := s.lomsClient.GetStock(ctx, sku)
	if err != nil {
		return fmt.Errorf("get stocks error: %w", err)
	}

	if uint64(count) > available {
		return fmt.Errorf(
			"insufficient stock, requested %d, got %d: %w",
			count,
			available,
			xerrors.ErrInsufficientStock,
		)
	}

	item := &entity.Item{
		SKU:   sku,
		Count: count,
	}

	if err := s.cartRepository.AddItem(ctx, userID, item); err != nil {
		return err
	}

	return nil
}
