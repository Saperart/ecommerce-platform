package product

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (p *productService) CreateProduct(ctx context.Context, name string, price uint32) (uint32, error) {
	if name == "" || price == 0 {
		return 0, xerrors.ErrInvalidInput
	}
	return p.productRepository.CreateProduct(ctx, name, price)
}
