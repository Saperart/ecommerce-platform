package product_test

import (
	"context"
	"testing"

	controllerproduct "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductServerGetProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *productpb.GetProductRequest
		setupMocks  func(productService *controllermocks.MockproductService)
		wantResp    *productpb.GetProductResponse
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &productpb.GetProductRequest{
				Sku: 10,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(&entity.Product{
						SKU:   10,
						Name:  "Кроссовки",
						Price: 49000,
					}, nil)
			},
			wantResp: &productpb.GetProductResponse{
				Name:  "Кроссовки",
				Price: 49000,
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &productpb.GetProductRequest{
				Sku: 0,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					GetProduct(gomock.Any(), uint32(0)).
					Return(nil, xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "product not found",
			req: &productpb.GetProductRequest{
				Sku: 404,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					GetProduct(gomock.Any(), uint32(404)).
					Return(nil, xerrors.ErrProductNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "unexpected internal error",
			req: &productpb.GetProductRequest{
				Sku: 12,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					GetProduct(gomock.Any(), uint32(12)).
					Return(nil, context.DeadlineExceeded)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			productService := controllermocks.NewMockproductService(ctrl)

			test.setupMocks(productService)

			server := controllerproduct.NewProductServer(productService)

			resp, err := server.GetProduct(context.Background(), test.req)

			if !test.wantErr {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, test.wantResp, resp)
				return
			}

			require.Error(t, err)
			if test.wantNilResp {
				require.Nil(t, resp)
			}

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, test.wantCode, st.Code())
		})
	}
}
