package item

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
)

//go:generate mockgen -source=item.go -destination=mocks/item_mock.go -package=mocks

type (
	cartRepository interface {
		AddItem(ctx context.Context, userID int64, item *entity.Item) error
		DeleteItem(ctx context.Context, userID int64, sku uint32) error
	}

	productClient interface {
		GetProduct(ctx context.Context, sku uint32) (*port.ProductInfo, error)
	}

	lomsClient interface {
		GetStock(ctx context.Context, sku uint32) (uint64, error)
	}
)

type itemService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewItemService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient) *itemService {
	return &itemService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
	}
}
