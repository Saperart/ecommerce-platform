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

func TestLOMSServiceCancelOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		orderID    int64
		setupMocks func(
			orderRepository *mocksusecase.MockorderRepository,
			stocksRepository *mocksusecase.MockstocksRepository,
		)
		wantErr error
	}{
		{
			name:       "invalid input",
			orderID:    0,
			setupMocks: func(_ *mocksusecase.MockorderRepository, _ *mocksusecase.MockstocksRepository) {},
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name:    "get order error",
			orderID: 404,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				_ *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(404)).
					Return(nil, xerrors.ErrOrderNotFound)
			},
			wantErr: xerrors.ErrOrderNotFound,
		},
		{
			name:    "invalid order status",
			orderID: 1002,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				_ *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(1002)).
					Return(&entity.Order{
						ID:     1002,
						Status: entity.OrderStatusPaid,
					}, nil)
			},
			wantErr: xerrors.ErrInvalidOrderStatus,
		},
		{
			name:    "set status error",
			orderID: 1004,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				_ *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(1004)).
					Return(&entity.Order{
						ID:     1004,
						Status: entity.OrderStatusAwaitingPayment,
						Items: []entity.OrderItem{
							{SKU: 10, Count: 2},
							{SKU: 12, Count: 1},
						},
					}, nil)

				orderRepository.EXPECT().
					SetOrderStatus(gomock.Any(), int64(1004), entity.OrderStatusCancelled).
					Return(context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:    "release stock error",
			orderID: 1003,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(1003)).
					Return(&entity.Order{
						ID:     1003,
						Status: entity.OrderStatusAwaitingPayment,
						Items: []entity.OrderItem{
							{SKU: 10, Count: 2},
							{SKU: 12, Count: 1},
						},
					}, nil)

				orderRepository.EXPECT().
					SetOrderStatus(gomock.Any(), int64(1003), entity.OrderStatusCancelled).
					Return(nil)

				stocksRepository.EXPECT().
					ReleaseStock(gomock.Any(), uint32(10), uint64(2)).
					Return(context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:    "success",
			orderID: 1001,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				stocksRepository *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(1001)).
					Return(&entity.Order{
						ID:     1001,
						Status: entity.OrderStatusAwaitingPayment,
						Items: []entity.OrderItem{
							{SKU: 10, Count: 2},
							{SKU: 12, Count: 1},
						},
					}, nil)

				orderRepository.EXPECT().
					SetOrderStatus(gomock.Any(), int64(1001), entity.OrderStatusCancelled).
					Return(nil)

				stocksRepository.EXPECT().
					ReleaseStock(gomock.Any(), uint32(10), uint64(2)).
					Return(nil)

				stocksRepository.EXPECT().
					ReleaseStock(gomock.Any(), uint32(12), uint64(1)).
					Return(nil)
			},
			wantErr: nil,
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

			err := service.CancelOrder(context.Background(), test.orderID)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
