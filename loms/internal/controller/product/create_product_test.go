package product_test

import (
	"context"
	"testing"

	controllerproduct "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductServerCreateProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *productpb.CreateProductRequest
		setupMocks  func(productService *controllermocks.MockproductService)
		wantResp    *productpb.CreateProductResponse
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &productpb.CreateProductRequest{
				Name:  "Майка",
				Price: 27000,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					CreateProduct(gomock.Any(), "Майка", uint32(27000)).
					Return(uint32(12), nil)
			},
			wantResp: &productpb.CreateProductResponse{
				Sku: 12,
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &productpb.CreateProductRequest{
				Name:  "",
				Price: 27000,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					CreateProduct(gomock.Any(), "", uint32(27000)).
					Return(uint32(0), xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "unexpected internal error",
			req: &productpb.CreateProductRequest{
				Name:  "Стул",
				Price: 11000,
			},
			setupMocks: func(productService *controllermocks.MockproductService) {
				productService.EXPECT().
					CreateProduct(gomock.Any(), "Стул", uint32(11000)).
					Return(uint32(0), context.DeadlineExceeded)
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

			resp, err := server.CreateProduct(context.Background(), test.req)

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
