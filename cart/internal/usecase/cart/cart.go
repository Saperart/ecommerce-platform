package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mock.go -package=mocks

type (
	cartRepository interface {
		GetItemsByUserID(ctx context.Context, userID int64) ([]*entity.Item, error)
		DeleteItemsByUserID(ctx context.Context, userID int64) error
	}

	productClient interface {
		GetProduct(ctx context.Context, sku uint32) (*port.ProductInfo, error)
	}

	lomsClient interface {
		CreateOrder(ctx context.Context, userID int64, items []*entity.Item) (int64, error)
	}
)

type cartService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewCartService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient) *cartService {
	return &cartService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
	}
}
