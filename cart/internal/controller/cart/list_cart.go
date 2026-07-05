package cart

import (
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *cartServer) ListCart(req *cartpb.ListCartRequest, srv cartpb.Cart_ListCartServer) error {
	if err := req.ValidateAll(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	items, totalPrice, err := s.cartService.ListCart(srv.Context(), req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, xerrors.ErrProductNotFound):
			return status.Errorf(codes.NotFound, "product not found: %v", err)
		default:
			return status.Error(codes.Internal, "internal error")
		}
	}
	if len(items) == 0 {
		return nil
	}

	responseItems := make([]*cartpb.Item, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, portItemToProto(item))
	}
	return srv.Send(&cartpb.ListCartResponse{
		Items:      responseItems,
		TotalPrice: totalPrice,
	})
}
