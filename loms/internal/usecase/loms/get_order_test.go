package loms_test

import (
	"context"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLOMSServiceGetOrder(t *testing.T) {
	t.Parallel()

	order := &entity.Order{
		ID:     1001,
		UserID: 42,
		Status: entity.OrderStatusAwaitingPayment,
		Items: []entity.OrderItem{
			{SKU: 10, Count: 2},
			{SKU: 12, Count: 1},
		},
		CreatedAt: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		orderID    int64
		setupMocks func(
			orderRepository *mocksusecase.MockorderRepository,
			stocksRepository *mocksusecase.MockstocksRepository,
		)
		wantOrder *entity.Order
		wantErr   error
	}{
		{
			name:       "invalid input",
			orderID:    0,
			setupMocks: func(_ *mocksusecase.MockorderRepository, _ *mocksusecase.MockstocksRepository) {},
			wantOrder:  nil,
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name:    "repository error",
			orderID: 404,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				_ *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(404)).
					Return(nil, xerrors.ErrOrderNotFound)
			},
			wantOrder: nil,
			wantErr:   xerrors.ErrOrderNotFound,
		},
		{
			name:    "success",
			orderID: 1001,
			setupMocks: func(
				orderRepository *mocksusecase.MockorderRepository,
				_ *mocksusecase.MockstocksRepository,
			) {
				orderRepository.EXPECT().
					GetOrder(gomock.Any(), int64(1001)).
					Return(order, nil)
			},
			wantOrder: order,
			wantErr:   nil,
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

			gotOrder, err := service.GetOrder(context.Background(), test.orderID)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Nil(t, gotOrder)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantOrder, gotOrder)
		})
	}
}
