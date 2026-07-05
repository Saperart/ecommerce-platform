package stocks

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

func (s *stocksService) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	if sku == 0 {
		return 0, xerrors.ErrInvalidInput
	}
	return s.stocksRepository.GetStock(ctx, sku)
}
