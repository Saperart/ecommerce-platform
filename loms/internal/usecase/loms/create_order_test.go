package loms_test

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLOMSServiceCreateOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		items      []entity.OrderItem
		setupMocks func(
			orderRepository *mocksusecase.MockorderRepository,
			stocksRepository *mocksusecase.MockstocksRepository,
		)
		wantOrderID int64
		wantErr     error
	}{
		{
			name:        "invalid input empty user id",
			userID:      0,
			items:       []entity.OrderItem{{SKU: 10, Count: 1}},
			setupMocks:  func(_ *mocksusecase.MockorderRepository, _ *mocksusecase.MockstocksRepository) {},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInvalidInput,
		},
		{
			name:        "invalid input empty items",
			userID:      42,
			items:       []entity.OrderItem{},
			setupMocks:  func(_ *mocksusecase.MockorderRepository, _ *mocksusecase.MockstocksRepository) {},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInvalidInput,
		},
		{
			name:   "invalid input zero sku",
			userID: 42,
			items: []entity.OrderItem{
				{SKU: 0, Count: 1},
			},
			setupMocks:  func(_ *mocksusecase.MockorderRepository, _ *mocksusecase.MockstocksRepository) {},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInvalidInput,
		},
		{
			name:   "invalid input zero count",
			userID: 42,
			items: []entity.OrderItem{
				{SKU: 10, Count: 0},
			},
			setupMocks:  func(_ *mocksusecase.MockorderRepository, _ *mocksusecase.MockstocksRepository) {},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInvalidInput,
		},
		{
			name:   "reserve stocks error",
			userID: 42,
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
			},
			setupMocks: func(
				_ *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
			) {
				stocksRepository.EXPECT().
					ReserveStocks(gomock.Any(), []entity.OrderItem{
						{SKU: 10, Count: 2},
					}).
					Return(xerrors.ErrInsufficientStock)
			},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInsufficientStock,
		},
		{
			name:   "create order repository error",
			userID: 42,
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 12, Count: 1},
			},
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
			) {
				items := []entity.OrderItem{
					{SKU: 10, Count: 2},
					{SKU: 12, Count: 1},
				}

				stocksRepository.EXPECT().
					ReserveStocks(gomock.Any(), items).
					Return(nil)

				orderRepository.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(0), context.DeadlineExceeded)
			},
			wantOrderID: 0,
			wantErr:     context.DeadlineExceeded,
		},
		{
			name:   "success",
			userID: 42,
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 12, Count: 1},
			},
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
			) {
				items := []entity.OrderItem{
					{SKU: 10, Count: 2},
					{SKU: 12, Count: 1},
				}

				stocksRepository.EXPECT().
					ReserveStocks(gomock.Any(), items).
					Return(nil)

				orderRepository.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(1001), nil)
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

			test.setupMocks(orderRepository, stocksRepository)

			service := lomsusecase.NewLomsService(orderRepository, stocksRepository)

			orderID, err := service.CreateOrder(context.Background(), test.userID, test.items)

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
