package stocks

import (
	"context"
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *stocksServer) GetStock(
	ctx context.Context,
	req *stockspb.GetStockRequest,
) (*stockspb.GetStockResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	count, err := s.stocksService.GetStock(ctx, req.GetSku())
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, xerrors.ErrStockNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &stockspb.GetStockResponse{Count: count}, nil
}
