package grpc

import (
	"context"
	"testing"

	adaptermocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/loms/grpc/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      int64
		items       []*entity.Item
		setupMocks  func(client *adaptermocks.MockLomsClient)
		wantOrderID int64
		wantErr     error
	}{
		{
			name:   "success",
			userID: 42,
			items: []*entity.Item{
				{SKU: 10, Count: 2},
				nil,
				{SKU: 12, Count: 1},
			},
			setupMocks: func(client *adaptermocks.MockLomsClient) {
				client.EXPECT().
					CreateOrder(
						gomock.Any(),
						&lomspb.CreateOrderRequest{
							UserId: 42,
							Items: []*lomspb.Item{
								{Sku: 10, Count: 2},
								{Sku: 12, Count: 1},
							},
						},
					).
					Return(&lomspb.CreateOrderResponse{
						OrderId: 1001,
					}, nil)
			},
			wantOrderID: 1001,
			wantErr:     nil,
		},
		{
			name:   "invalid input",
			userID: 0,
			items: []*entity.Item{
				{SKU: 10, Count: 1},
			},
			setupMocks: func(client *adaptermocks.MockLomsClient) {
				client.EXPECT().
					CreateOrder(
						gomock.Any(),
						&lomspb.CreateOrderRequest{
							UserId: 0,
							Items: []*lomspb.Item{
								{Sku: 10, Count: 1},
							},
						},
					).
					Return(nil, status.Error(codes.InvalidArgument, "invalid input"))
			},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInvalidInput,
		},
		{
			name:   "insufficient stock",
			userID: 42,
			items: []*entity.Item{
				{SKU: 10, Count: 100},
			},
			setupMocks: func(client *adaptermocks.MockLomsClient) {
				client.EXPECT().
					CreateOrder(
						gomock.Any(),
						&lomspb.CreateOrderRequest{
							UserId: 42,
							Items: []*lomspb.Item{
								{Sku: 10, Count: 100},
							},
						},
					).
					Return(nil, status.Error(codes.FailedPrecondition, "insufficient stock"))
			},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInsufficientStock,
		},
		{
			name:   "unexpected grpc error",
			userID: 42,
			items: []*entity.Item{
				{SKU: 10, Count: 1},
			},
			setupMocks: func(client *adaptermocks.MockLomsClient) {
				client.EXPECT().
					CreateOrder(
						gomock.Any(),
						&lomspb.CreateOrderRequest{
							UserId: 42,
							Items: []*lomspb.Item{
								{Sku: 10, Count: 1},
							},
						},
					).
					Return(nil, context.DeadlineExceeded)
			},
			wantOrderID: 0,
			wantErr:     context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			stocksClient := adaptermocks.NewMockStocksClient(ctrl)
			lomsClient := adaptermocks.NewMockLomsClient(ctrl)

			test.setupMocks(lomsClient)

			adapter := NewLOMSClient(stocksClient, lomsClient)

			orderID, err := adapter.CreateOrder(context.Background(), test.userID, test.items)

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
