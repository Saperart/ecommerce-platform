package cart

import (
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
)

func portItemToProto(item *port.Item) *cartpb.Item {
	if item == nil {
		return nil
	}
	return &cartpb.Item{
		Sku:   item.SKU,
		Count: item.Count,
		Name:  item.Name,
		Price: item.Price,
	}
}
