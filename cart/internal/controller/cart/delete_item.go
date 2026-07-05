package cart

import (
	"context"
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *cartServer) DeleteItem(ctx context.Context, req *cartpb.DeleteItemRequest) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	if err := s.itemService.DeleteItem(ctx, req.GetUserId(), req.GetSku()); err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, xerrors.ErrItemNotFound) || errors.Is(err, xerrors.ErrCartNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}
