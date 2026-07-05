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

func TestCartServerDeleteItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *cartpb.DeleteItemRequest
		setupMocks  func(itemService *controllermocks.MockItemService)
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &cartpb.DeleteItemRequest{
				UserId: 42,
				Sku:    1001,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					DeleteItem(gomock.Any(), int64(42), uint32(1001)).
					Return(nil)
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input from service",
			req: &cartpb.DeleteItemRequest{
				UserId: 0,
				Sku:    1001,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					DeleteItem(gomock.Any(), int64(0), uint32(1001)).
					Return(xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "item not found",
			req: &cartpb.DeleteItemRequest{
				UserId: 7,
				Sku:    404,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					DeleteItem(gomock.Any(), int64(7), uint32(404)).
					Return(xerrors.ErrItemNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "cart not found",
			req: &cartpb.DeleteItemRequest{
				UserId: 99,
				Sku:    1001,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					DeleteItem(gomock.Any(), int64(99), uint32(1001)).
					Return(xerrors.ErrCartNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "unexpected internal error",
			req: &cartpb.DeleteItemRequest{
				UserId: 5,
				Sku:    1002,
			},
			setupMocks: func(itemService *controllermocks.MockItemService) {
				itemService.EXPECT().
					DeleteItem(gomock.Any(), int64(5), uint32(1002)).
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

			resp, err := server.DeleteItem(context.Background(), test.req)

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
