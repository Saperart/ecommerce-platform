package product

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

//go:generate mockgen -source=product.go -destination=mocks/product_mock.go -package=mocks

type (
	productRepository interface {
		GetProduct(ctx context.Context, sku uint32) (*entity.Product, error)
		CreateProduct(ctx context.Context, name string, price uint32) (uint32, error)
	}
)

type productService struct {
	productRepository productRepository
}

func NewProductService(productRepository productRepository) *productService {
	return &productService{
		productRepository: productRepository,
	}
}
