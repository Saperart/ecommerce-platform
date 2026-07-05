package loms_test

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLOMSServiceOrderStatusChangedNotificationKindHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		payload    []byte
		options    []lomsusecase.Option
		setupMocks func(notifications *mocksusecase.MocknotificationsClient)
		wantErr    bool
	}{
		{
			name:       "notifications client is not configured",
			payload:    []byte(`{`),
			options:    nil,
			setupMocks: func(_ *mocksusecase.MocknotificationsClient) {},
			wantErr:    false,
		},
		{
			name:       "invalid payload",
			payload:    []byte(`{`),
			options:    []lomsusecase.Option{},
			setupMocks: func(_ *mocksusecase.MocknotificationsClient) {},
			wantErr:    true,
		},
		{
			name: "notifications error",
			payload: []byte(`{
				"user_id": 42,
				"order_id": 1001,
				"status": "paid"
			}`),
			options: []lomsusecase.Option{},
			setupMocks: func(notifications *mocksusecase.MocknotificationsClient) {
				notifications.EXPECT().
					SendOrderStatusChangedNotification(gomock.Any(), port.OrderStatusChangedNotification{
						UserID:  42,
						OrderID: 1001,
						Status:  port.OrderStatusPaid,
					}).
					Return(context.DeadlineExceeded)
			},
			wantErr: true,
		},
		{
			name: "success",
			payload: []byte(`{
				"user_id": 42,
				"order_id": 1001,
				"status": "awaiting_payment"
			}`),
			options: []lomsusecase.Option{},
			setupMocks: func(notifications *mocksusecase.MocknotificationsClient) {
				notifications.EXPECT().
					SendOrderStatusChangedNotification(gomock.Any(), port.OrderStatusChangedNotification{
						UserID:  42,
						OrderID: 1001,
						Status:  port.OrderStatusAwaitingPayment,
					}).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			orderRepository := mocksusecase.NewMockorderRepository(ctrl)
			stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
			notifications := mocksusecase.NewMocknotificationsClient(ctrl)

			test.setupMocks(notifications)

			options := test.options
			if options != nil {
				options = append(options, lomsusecase.WithNotificationsClient(notifications))
			}

			service := lomsusecase.NewLomsService(orderRepository, stocksRepository, options...)

			err := service.OrderStatusChangedNotificationKindHandler(context.Background(), test.payload)

			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLOMSServiceCreateOrderSavesOutboxMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(
			orderRepository *mocksusecase.MockorderRepository,
			stocksRepository *mocksusecase.MockstocksRepository,
			outboxRepository *mocksusecase.MockoutboxRepository,
			transactor *mocksusecase.Mocktransactor,
		)
		wantOrderID int64
		wantErr     error
	}{
		{
			name: "save outbox error",
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
				outboxRepository *mocksusecase.MockoutboxRepository,
				transactor *mocksusecase.Mocktransactor,
			) {
				items := []entity.OrderItem{{SKU: 10, Count: 2}}

				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})
				stocksRepository.EXPECT().
					ReserveStocks(gomock.Any(), items).
					Return(nil)
				orderRepository.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(1001), nil)
				outboxRepository.EXPECT().
					SaveMessage(gomock.Any(), "order-status:1001:awaiting_payment", outboxrepo.KindNotification, gomock.Any()).
					Return(context.DeadlineExceeded)
			},
			wantOrderID: 0,
			wantErr:     context.DeadlineExceeded,
		},
		{
			name: "success",
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
				outboxRepository *mocksusecase.MockoutboxRepository,
				transactor *mocksusecase.Mocktransactor,
			) {
				items := []entity.OrderItem{{SKU: 10, Count: 2}}

				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})
				stocksRepository.EXPECT().
					ReserveStocks(gomock.Any(), items).
					Return(nil)
				orderRepository.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(1001), nil)
				outboxRepository.EXPECT().
					SaveMessage(gomock.Any(), "order-status:1001:awaiting_payment", outboxrepo.KindNotification, gomock.Any()).
					Return(nil)
			},
			wantOrderID: 1001,
			wantErr:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			orderRepository := mocksusecase.NewMockorderRepository(ctrl)
			stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
			outboxRepository := mocksusecase.NewMockoutboxRepository(ctrl)
			transactor := mocksusecase.NewMocktransactor(ctrl)

			test.setupMocks(orderRepository, stocksRepository, outboxRepository, transactor)

			service := lomsusecase.NewLomsService(
				orderRepository,
				stocksRepository,
				lomsusecase.WithOutboxRepository(outboxRepository),
				lomsusecase.WithTransactor(transactor),
			)

			orderID, err := service.CreateOrder(context.Background(), 42, []entity.OrderItem{{SKU: 10, Count: 2}})

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantOrderID, orderID)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantOrderID, orderID)
		})
	}
}

func TestLOMSServicePayOrderSavesOutboxMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	orderRepository := mocksusecase.NewMockorderRepository(ctrl)
	stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
	outboxRepository := mocksusecase.NewMockoutboxRepository(ctrl)
	transactor := mocksusecase.NewMocktransactor(ctrl)

	orderRepository.EXPECT().
		GetOrder(gomock.Any(), int64(1001)).
		Return(&entity.Order{
			ID:     1001,
			UserID: 42,
			Status: entity.OrderStatusAwaitingPayment,
		}, nil)
	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})
	orderRepository.EXPECT().
		SetOrderStatus(gomock.Any(), int64(1001), entity.OrderStatusPaid).
		Return(nil)
	outboxRepository.EXPECT().
		SaveMessage(gomock.Any(), "order-status:1001:paid", outboxrepo.KindNotification, gomock.Any()).
		Return(nil)

	service := lomsusecase.NewLomsService(
		orderRepository,
		stocksRepository,
		lomsusecase.WithOutboxRepository(outboxRepository),
		lomsusecase.WithTransactor(transactor),
	)

	err := service.PayOrder(context.Background(), 1001)

	require.NoError(t, err)
}

func TestLOMSServicePayOrderOutboxError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	orderRepository := mocksusecase.NewMockorderRepository(ctrl)
	stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
	outboxRepository := mocksusecase.NewMockoutboxRepository(ctrl)
	transactor := mocksusecase.NewMocktransactor(ctrl)

	orderRepository.EXPECT().
		GetOrder(gomock.Any(), int64(1001)).
		Return(&entity.Order{
			ID:     1001,
			UserID: 42,
			Status: entity.OrderStatusAwaitingPayment,
		}, nil)
	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})
	orderRepository.EXPECT().
		SetOrderStatus(gomock.Any(), int64(1001), entity.OrderStatusPaid).
		Return(nil)
	outboxRepository.EXPECT().
		SaveMessage(gomock.Any(), "order-status:1001:paid", outboxrepo.KindNotification, gomock.Any()).
		Return(context.DeadlineExceeded)

	service := lomsusecase.NewLomsService(
		orderRepository,
		stocksRepository,
		lomsusecase.WithOutboxRepository(outboxRepository),
		lomsusecase.WithTransactor(transactor),
	)

	err := service.PayOrder(context.Background(), 1001)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestLOMSServicePayOrderTransactorError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	orderRepository := mocksusecase.NewMockorderRepository(ctrl)
	stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
	transactor := mocksusecase.NewMocktransactor(ctrl)

	orderRepository.EXPECT().
		GetOrder(gomock.Any(), int64(1001)).
		Return(&entity.Order{
			ID:     1001,
			UserID: 42,
			Status: entity.OrderStatusAwaitingPayment,
		}, nil)
	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		Return(context.DeadlineExceeded)

	service := lomsusecase.NewLomsService(
		orderRepository,
		stocksRepository,
		lomsusecase.WithTransactor(transactor),
	)

	err := service.PayOrder(context.Background(), 1001)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestLOMSServiceCancelOrderSavesOutboxMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	orderRepository := mocksusecase.NewMockorderRepository(ctrl)
	stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
	outboxRepository := mocksusecase.NewMockoutboxRepository(ctrl)
	transactor := mocksusecase.NewMocktransactor(ctrl)

	orderRepository.EXPECT().
		GetOrder(gomock.Any(), int64(1001)).
		Return(&entity.Order{
			ID:     1001,
			UserID: 42,
			Status: entity.OrderStatusAwaitingPayment,
			Items:  []entity.OrderItem{{SKU: 10, Count: 2}},
		}, nil)
	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})
	orderRepository.EXPECT().
		SetOrderStatus(gomock.Any(), int64(1001), entity.OrderStatusCancelled).
		Return(nil)
	stocksRepository.EXPECT().
		ReleaseStock(gomock.Any(), uint32(10), uint64(2)).
		Return(nil)
	outboxRepository.EXPECT().
		SaveMessage(gomock.Any(), "order-status:1001:cancelled", outboxrepo.KindNotification, gomock.Any()).
		Return(nil)

	service := lomsusecase.NewLomsService(
		orderRepository,
		stocksRepository,
		lomsusecase.WithOutboxRepository(outboxRepository),
		lomsusecase.WithTransactor(transactor),
	)

	err := service.CancelOrder(context.Background(), 1001)

	require.NoError(t, err)
}

func TestLOMSServiceCancelOrderOutboxError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	orderRepository := mocksusecase.NewMockorderRepository(ctrl)
	stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)
	outboxRepository := mocksusecase.NewMockoutboxRepository(ctrl)
	transactor := mocksusecase.NewMocktransactor(ctrl)

	orderRepository.EXPECT().
		GetOrder(gomock.Any(), int64(1001)).
		Return(&entity.Order{
			ID:     1001,
			UserID: 42,
			Status: entity.OrderStatusAwaitingPayment,
			Items:  []entity.OrderItem{{SKU: 10, Count: 2}},
		}, nil)
	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})
	orderRepository.EXPECT().
		SetOrderStatus(gomock.Any(), int64(1001), entity.OrderStatusCancelled).
		Return(nil)
	stocksRepository.EXPECT().
		ReleaseStock(gomock.Any(), uint32(10), uint64(2)).
		Return(nil)
	outboxRepository.EXPECT().
		SaveMessage(gomock.Any(), "order-status:1001:cancelled", outboxrepo.KindNotification, gomock.Any()).
		Return(context.DeadlineExceeded)

	service := lomsusecase.NewLomsService(
		orderRepository,
		stocksRepository,
		lomsusecase.WithOutboxRepository(outboxRepository),
		lomsusecase.WithTransactor(transactor),
	)

	err := service.CancelOrder(context.Background(), 1001)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
