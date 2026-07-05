package grpc

import (
	"context"
	"testing"

	adaptermocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/product/grpc/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductClient_GetProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sku        uint32
		setupMocks func(client *adaptermocks.MockProductServiceClient)
		wantResult *port.ProductInfo
		wantErr    error
	}{
		{
			name: "success",
			sku:  10,
			setupMocks: func(client *adaptermocks.MockProductServiceClient) {
				client.EXPECT().
					GetProduct(gomock.Any(), &productpb.GetProductRequest{Sku: 10}).
					Return(&productpb.GetProductResponse{
						Name:  "Кроссовки",
						Price: 49000,
					}, nil)
			},
			wantResult: &port.ProductInfo{
				Name:  "Кроссовки",
				Price: 49000,
			},
			wantErr: nil,
		},
		{
			name: "product not found",
			sku:  404,
			setupMocks: func(client *adaptermocks.MockProductServiceClient) {
				client.EXPECT().
					GetProduct(gomock.Any(), &productpb.GetProductRequest{Sku: 404}).
					Return(nil, status.Error(codes.NotFound, "product not found"))
			},
			wantResult: nil,
			wantErr:    xerrors.ErrProductNotFound,
		},
		{
			name: "invalid input",
			sku:  0,
			setupMocks: func(client *adaptermocks.MockProductServiceClient) {
				client.EXPECT().
					GetProduct(gomock.Any(), &productpb.GetProductRequest{Sku: 0}).
					Return(nil, status.Error(codes.InvalidArgument, "invalid sku"))
			},
			wantResult: nil,
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name: "unexpected grpc error",
			sku:  10,
			setupMocks: func(client *adaptermocks.MockProductServiceClient) {
				client.EXPECT().
					GetProduct(gomock.Any(), &productpb.GetProductRequest{Sku: 10}).
					Return(nil, context.DeadlineExceeded)
			},
			wantResult: nil,
			wantErr:    context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			client := adaptermocks.NewMockProductServiceClient(ctrl)
			test.setupMocks(client)

			adapter := NewProductClient(client)

			gotResult, err := adapter.GetProduct(context.Background(), test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Nil(t, gotResult)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}
