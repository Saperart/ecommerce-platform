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

func TestCartServerCheckoutCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *cartpb.CheckoutCartRequest
		setupMocks  func(cartService *controllermocks.MockCartService)
		wantOrderID int64
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &cartpb.CheckoutCartRequest{
				UserId: 42,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					CheckoutCart(gomock.Any(), int64(42)).
					Return(int64(1001), nil)
			},
			wantOrderID: 1001,
		},
		{
			name: "invalid input",
			req: &cartpb.CheckoutCartRequest{
				UserId: 0,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					CheckoutCart(gomock.Any(), int64(0)).
					Return(int64(0), xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			itemService := controllermocks.NewMockItemService(ctrl)
			cartService := controllermocks.NewMockCartService(ctrl)

			test.setupMocks(cartService)

			server := controllercart.NewCartServer(itemService, cartService)

			resp, err := server.CheckoutCart(context.Background(), test.req)

			if !test.wantErr {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, test.wantOrderID, resp.GetOrderId())
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
