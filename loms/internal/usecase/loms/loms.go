package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres"
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mock.go -package=mocks

type (
	orderRepository interface {
		CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error)
		DeleteOrder(ctx context.Context, orderID int64) error
		GetOrder(ctx context.Context, orderID int64) (*entity.Order, error)
		SetOrderStatus(ctx context.Context, orderID int64, status entity.OrderStatus) error
	}

	stocksRepository interface {
		GetStock(ctx context.Context, sku uint32) (uint64, error)
		ReserveStocks(ctx context.Context, items []entity.OrderItem) error
		ReleaseStock(ctx context.Context, sku uint32, count uint64) error
	}

	transactor interface {
		WithTx(ctx context.Context, f func(ctx context.Context) error) error
	}

	outboxRepository interface {
		SaveMessage(ctx context.Context, idempotencyKey string, kind outboxrepo.Kind, payload []byte) error
	}

	notificationsClient interface {
		SendOrderStatusChangedNotification(ctx context.Context, notification port.OrderStatusChangedNotification) error
	}
)

type lomsService struct {
	orderRepository  orderRepository
	stocksRepository stocksRepository
	transactor       transactor
	outboxRepository outboxRepository
	notifications    notificationsClient
}

type Option func(*lomsService)

func WithTransactor(transactor transactor) Option {
	return func(s *lomsService) {
		s.transactor = transactor
	}
}

func WithOutboxRepository(outboxRepository outboxRepository) Option {
	return func(s *lomsService) {
		s.outboxRepository = outboxRepository
	}
}

func WithNotificationsClient(notifications notificationsClient) Option {
	return func(s *lomsService) {
		s.notifications = notifications
	}
}

func NewLomsService(
	orderRepository orderRepository,
	stocksRepository stocksRepository,
	options ...Option,
) *lomsService {
	service := &lomsService{
		orderRepository:  orderRepository,
		stocksRepository: stocksRepository,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *lomsService) withTx(ctx context.Context, f func(ctx context.Context) error) error {
	if s.transactor == nil {
		return f(ctx)
	}
	return s.transactor.WithTx(ctx, f)
}
