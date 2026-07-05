package product

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
)

//go:generate mockgen -source=product.go -destination=mocks/product_mock.go -package=mocks

type (
	productService interface {
		GetProduct(ctx context.Context, sku uint32) (*entity.Product, error)
		CreateProduct(ctx context.Context, name string, price uint32) (uint32, error)
	}
)

type productServer struct {
	productpb.UnimplementedProductServiceServer
	productService productService
}

func NewProductServer(productService productService) *productServer {
	return &productServer{
		productService: productService,
	}
}
