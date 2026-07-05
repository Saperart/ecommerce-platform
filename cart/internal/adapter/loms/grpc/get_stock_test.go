package grpc

import (
	"context"
	"testing"

	adaptermocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/loms/grpc/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLOMSClient_GetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sku        uint32
		setupMocks func(client *adaptermocks.MockStocksClient)
		wantCount  uint64
		wantErr    error
	}{
		{
			name: "success",
			sku:  10,
			setupMocks: func(client *adaptermocks.MockStocksClient) {
				client.EXPECT().
					GetStock(gomock.Any(), &stockspb.GetStockRequest{Sku: 10}).
					Return(&stockspb.GetStockResponse{
						Count: 15,
					}, nil)
			},
			wantCount: 15,
			wantErr:   nil,
		},
		{
			name: "invalid input",
			sku:  0,
			setupMocks: func(client *adaptermocks.MockStocksClient) {
				client.EXPECT().
					GetStock(gomock.Any(), &stockspb.GetStockRequest{Sku: 0}).
					Return(nil, status.Error(codes.InvalidArgument, "invalid sku"))
			},
			wantCount: 0,
			wantErr:   xerrors.ErrInvalidInput,
		},
		{
			name: "stock not found",
			sku:  404,
			setupMocks: func(client *adaptermocks.MockStocksClient) {
				client.EXPECT().
					GetStock(gomock.Any(), &stockspb.GetStockRequest{Sku: 404}).
					Return(nil, status.Error(codes.NotFound, "stock not found"))
			},
			wantCount: 0,
			wantErr:   xerrors.ErrStockNotFound,
		},
		{
			name: "unexpected grpc error",
			sku:  10,
			setupMocks: func(client *adaptermocks.MockStocksClient) {
				client.EXPECT().
					GetStock(gomock.Any(), &stockspb.GetStockRequest{Sku: 10}).
					Return(nil, context.DeadlineExceeded)
			},
			wantCount: 0,
			wantErr:   context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			stocksClient := adaptermocks.NewMockStocksClient(ctrl)
			lomsClient := adaptermocks.NewMockLomsClient(ctrl)

			test.setupMocks(stocksClient)

			adapter := NewLOMSClient(stocksClient, lomsClient)

			count, err := adapter.GetStock(context.Background(), test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantCount, count)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantCount, count)
		})
	}
}
