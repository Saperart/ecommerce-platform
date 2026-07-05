package cart_test

import (
	"context"
	"testing"

	controllercart "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCartServerAddItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *cartpb.AddItemRequest
		setupMocks  func(itemService *controllermocks.MockItemService)
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &cartpb.AddItemRequest{
				UserId: 42,
				Sku:    1001,
				Count:  2,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					AddItem(gomock.Any(), int64(42), uint32(1001), uint32(2)).
					Return(nil)
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input from service",
			req: &cartpb.AddItemRequest{
				UserId: 10,
				Sku:    1004,
				Count:  1,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					AddItem(gomock.Any(), int64(10), uint32(1004), uint32(1)).
					Return(xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "product not found",
			req: &cartpb.AddItemRequest{
				UserId: 1,
				Sku:    404,
				Count:  1,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					AddItem(gomock.Any(), int64(1), uint32(404), uint32(1)).
					Return(xerrors.ErrProductNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "insufficient stock",
			req: &cartpb.AddItemRequest{
				UserId: 5,
				Sku:    1002,
				Count:  999,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					AddItem(gomock.Any(), int64(5), uint32(1002), uint32(999)).
					Return(xerrors.ErrInsufficientStock)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.FailedPrecondition,
		},
		{
			name: "unexpected internal error",
			req: &cartpb.AddItemRequest{
				UserId: 3,
				Sku:    1003,
				Count:  1,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					AddItem(gomock.Any(), int64(3), uint32(1003), uint32(1)).
					Return(context.DeadlineExceeded)
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

			itemService := controllermocks.NewMockItemService(ctrl)
			cartService := controllermocks.NewMockCartService(ctrl)

			test.setupMocks(itemService)

			server := controllercart.NewCartServer(itemService, cartService)

			resp, err := server.AddItem(context.Background(), test.req)

			if !test.wantErr {
				require.NoError(t, err)
				require.NotNil(t, resp)
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
