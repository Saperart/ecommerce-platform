package cart

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
)

func (s *cartService) ListCart(ctx context.Context, userID int64) (items []*port.Item, totalPrice uint32, err error) {
	if userID <= 0 {
		return nil, 0, xerrors.ErrInvalidInput
	}

	allItems, err := s.cartRepository.GetItemsByUserID(ctx, userID)

	if err != nil {
		return nil, 0, err
	}

	totalPrice = 0
	items = make([]*port.Item, 0, len(allItems))
	for _, item := range allItems {
		productInfo, err := s.productClient.GetProduct(ctx, item.SKU)
		if err != nil {
			return nil, 0, err
		}
		finalItem := &port.Item{
			SKU:   item.SKU,
			Count: item.Count,
			Name:  productInfo.Name,
			Price: productInfo.Price,
		}
		totalPrice += productInfo.Price * item.Count
		items = append(items, finalItem)
	}
	return items, totalPrice, nil
}
