package stocks

import (
	"context"
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *stocksServer) SetStock(ctx context.Context, req *stockspb.SetStockRequest) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.stocksService.SetStock(ctx, req.GetSku(), req.GetCount()); err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}
