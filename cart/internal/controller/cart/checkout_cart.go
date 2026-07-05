package cart

import (
	"context"
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *cartServer) CheckoutCart(ctx context.Context, req *cartpb.CheckoutCartRequest) (*cartpb.CheckoutCartResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	orderID, err := s.cartService.CheckoutCart(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
		case errors.Is(err, xerrors.ErrCartIsEmpty) || errors.Is(err, xerrors.ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &cartpb.CheckoutCartResponse{OrderId: orderID}, nil
}
