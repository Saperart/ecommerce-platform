package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mock.go -package=mocks

type (
	ItemService interface {
		AddItem(ctx context.Context, userID int64, sku, count uint32) error
		DeleteItem(ctx context.Context, userID int64, sku uint32) error
	}

	//nolint:revive // Existing public controller contract; renaming churns generated mocks/tests.
	CartService interface {
		ListCart(ctx context.Context, userID int64) ([]*port.Item, uint32, error)
		CheckoutCart(ctx context.Context, userID int64) (int64, error)
		ClearCart(ctx context.Context, userID int64) error
	}
)

type cartServer struct {
	cartpb.UnimplementedCartServer
	itemService ItemService
	cartService CartService
}

func NewCartServer(
	itemService ItemService,
	cartService CartService,
) *cartServer {
	return &cartServer{
		itemService: itemService,
		cartService: cartService,
	}
}
