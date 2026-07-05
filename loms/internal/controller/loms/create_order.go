package loms

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/converter"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *lomsServer) CreateOrder(
	ctx context.Context,
	req *lomspb.CreateOrderRequest,
) (*lomspb.CreateOrderResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	orderID, err := s.lomsService.CreateOrder(ctx, req.GetUserId(), converter.ProtoToItemsOrder(req.GetItems()))
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		case errors.Is(err, xerrors.ErrStockNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &lomspb.CreateOrderResponse{OrderId: orderID}, nil
}
