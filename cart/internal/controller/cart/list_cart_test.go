package cart_test

import (
	"context"
	"testing"

	controllercart "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCartServerListCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		req               *cartpb.ListCartRequest
		setupMocks        func(cartService *controllermocks.MockCartService)
		wantCode          codes.Code
		wantErr           bool
		wantSentItems     []*cartpb.Item
		wantSentTotal     uint32
		wantNilStreamResp bool
	}{
		{
			name: "success",
			req: &cartpb.ListCartRequest{
				UserId: 42,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ListCart(gomock.Any(), int64(42)).
					Return([]*port.Item{
						{
							SKU:   10,
							Count: 2,
							Name:  "Кроссовки",
							Price: 49000,
						},
						{
							SKU:   12,
							Count: 1,
							Name:  "Майка",
							Price: 27000,
						},
						{
							SKU:   13,
							Count: 1,
							Name:  "Стул",
							Price: 11000,
						},
					}, uint32(136000), nil)
			},
			wantErr: false,
			wantSentItems: []*cartpb.Item{
				{
					Sku:   10,
					Count: 2,
					Name:  "Кроссовки",
					Price: 49000,
				},
				{
					Sku:   12,
					Count: 1,
					Name:  "Майка",
					Price: 27000,
				},
				{
					Sku:   13,
					Count: 1,
					Name:  "Стул",
					Price: 11000,
				},
			},
			wantSentTotal: 136000,
		},
		{
			name: "invalid input from service",
			req: &cartpb.ListCartRequest{
				UserId: 0,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ListCart(gomock.Any(), int64(0)).
					Return(nil, uint32(0), xerrors.ErrInvalidInput)
			},
			wantErr:           true,
			wantCode:          codes.InvalidArgument,
			wantNilStreamResp: true,
		},
		{
			name: "product not found",
			req: &cartpb.ListCartRequest{
				UserId: 7,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ListCart(gomock.Any(), int64(7)).
					Return(nil, uint32(0), xerrors.ErrProductNotFound)
			},
			wantErr:           true,
			wantCode:          codes.NotFound,
			wantNilStreamResp: true,
		},
		{
			name: "unexpected internal error",
			req: &cartpb.ListCartRequest{
				UserId: 5,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ListCart(gomock.Any(), int64(5)).
					Return(nil, uint32(0), context.DeadlineExceeded)
			},
			wantErr:           true,
			wantCode:          codes.Internal,
			wantNilStreamResp: true,
		},
		{
			name: "empty cart",
			req: &cartpb.ListCartRequest{
				UserId: 42,
			},
			setupMocks: func(cartService *controllermocks.MockCartService) {
				cartService.EXPECT().
					ListCart(gomock.Any(), int64(42)).
					Return(nil, uint32(0), nil)
			},
			wantErr:           false,
			wantNilStreamResp: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			itemService := controllermocks.NewMockItemService(ctrl)
			cartService := controllermocks.NewMockCartService(ctrl)
			stream := controllermocks.NewMockCart_ListCartServer(ctrl)
			var streamResponse *cartpb.ListCartResponse

			test.setupMocks(cartService)
			stream.EXPECT().
				Context().
				Return(context.Background())

			if !test.wantNilStreamResp {
				stream.EXPECT().
					Send(gomock.Any()).
					DoAndReturn(func(resp *cartpb.ListCartResponse) error {
						streamResponse = resp
						return nil
					})
			}

			server := controllercart.NewCartServer(itemService, cartService)

			err := server.ListCart(test.req, stream)

			if !test.wantErr {
				require.NoError(t, err)
				if test.wantNilStreamResp {
					require.Nil(t, streamResponse)
					return
				}
				require.NotNil(t, streamResponse)
				require.Equal(t, test.wantSentTotal, streamResponse.GetTotalPrice())
				require.Equal(t, test.wantSentItems, streamResponse.GetItems())
				return
			}

			require.Error(t, err)

			if test.wantNilStreamResp {
				require.Nil(t, streamResponse)
			}

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, test.wantCode, st.Code())
		})
	}
}
