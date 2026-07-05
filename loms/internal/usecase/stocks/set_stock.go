package stocks

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (s *stocksService) SetStock(ctx context.Context, sku uint32, count uint64) error {
	if sku == 0 {
		return xerrors.ErrInvalidInput
	}
	return s.stocksRepository.SetStock(ctx, sku, count)
}
