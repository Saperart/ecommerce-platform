package grpc

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *lomsClient) CreateOrder(ctx context.Context, userID int64, items []*entity.Item) (int64, error) {
	protoItems := make([]*lomspb.Item, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		protoItems = append(protoItems, &lomspb.Item{
			Sku:   item.SKU,
			Count: item.Count,
		})
	}

	response, err := s.lomsClient.CreateOrder(ctx, &lomspb.CreateOrderRequest{
		UserId: userID,
		Items:  protoItems,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			return 0, xerrors.ErrInvalidInput
		case codes.FailedPrecondition:
			return 0, xerrors.ErrInsufficientStock
		default:
			return 0, err
		}
	}
	return response.GetOrderId(), nil
}
