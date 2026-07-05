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

func TestCartServerClearCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *cartpb.ClearCartRequest
		setupMocks  func(cartService *controllermocks.MockCartService)
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &cartpb.ClearCartRequest{
				UserId: 42,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ClearCart(gomock.Any(), int64(42)).
					Return(nil)
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input from service",
			req: &cartpb.ClearCartRequest{
				UserId: 0,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ClearCart(gomock.Any(), int64(0)).
					Return(xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "unexpected internal error",
			req: &cartpb.ClearCartRequest{
				UserId: 5,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ClearCart(gomock.Any(), int64(5)).
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

			test.setupMocks(cartService)

			server := controllercart.NewCartServer(itemService, cartService)

			resp, err := server.ClearCart(context.Background(), test.req)

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
