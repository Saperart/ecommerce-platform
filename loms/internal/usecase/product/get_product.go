package product

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (p *productService) GetProduct(ctx context.Context, sku uint32) (*entity.Product, error) {
	if sku == 0 {
		return nil, xerrors.ErrInvalidInput
	}
	return p.productRepository.GetProduct(ctx, sku)
}
