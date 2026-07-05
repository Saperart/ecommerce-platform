package grpc

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *lomsClient) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	response, err := s.stocksClient.GetStock(ctx, &stockspb.GetStockRequest{Sku: sku})
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			return 0, xerrors.ErrInvalidInput
		case codes.NotFound:
			return 0, xerrors.ErrStockNotFound
		default:
			return 0, err
		}
	}
	return response.GetCount(), nil
}
