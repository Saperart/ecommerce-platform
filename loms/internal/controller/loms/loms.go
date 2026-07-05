package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mock.go -package=mocks

type (
	lomsService interface {
		CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error)
		PayOrder(ctx context.Context, orderID int64) error
		CancelOrder(ctx context.Context, orderID int64) error
		GetOrder(ctx context.Context, orderID int64) (*entity.Order, error)
	}
)

type lomsServer struct {
	lomspb.UnimplementedLomsServer
	lomsService lomsService
}

func NewLomsServer(lomsService lomsService) *lomsServer {
	return &lomsServer{
		lomsService: lomsService,
	}
}
