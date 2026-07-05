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

func (s *lomsServer) GetOrder(ctx context.Context, req *lomspb.GetOrderRequest) (*lomspb.GetOrderResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	order, err := s.lomsService.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, xerrors.ErrOrderNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return converter.OrderToProto(order), nil
}
