package loms_test

import (
	"context"
	"testing"

	controllerloms "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLOMSServerCreateOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *lomspb.CreateOrderRequest
		setupMocks  func(lomsService *controllermocks.MocklomsService)
		wantOrderID int64
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &lomspb.CreateOrderRequest{
				UserId: 42,
				Items: []*lomspb.Item{
					{Sku: 10, Count: 2},
					{Sku: 12, Count: 1},
				},
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					CreateOrder(
						gomock.Any(),
						int64(42),
						[]entity.OrderItem{
							{SKU: 10, Count: 2},
							{SKU: 12, Count: 1},
						},
					).
					Return(int64(1001), nil)
			},
			wantOrderID: 1001,
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &lomspb.CreateOrderRequest{
				UserId: 0,
				Items: []*lomspb.Item{
					{Sku: 10, Count: 1},
				},
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					CreateOrder(
						gomock.Any(),
						int64(0),
						[]entity.OrderItem{
							{SKU: 10, Count: 1},
						},
					).
					Return(int64(0), xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "stock not found",
			req: &lomspb.CreateOrderRequest{
				UserId: 42,
				Items: []*lomspb.Item{
					{Sku: 404, Count: 1},
				},
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					CreateOrder(
						gomock.Any(),
						int64(42),
						[]entity.OrderItem{
							{SKU: 404, Count: 1},
						},
					).
					Return(int64(0), xerrors.ErrStockNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "insufficient stock",
			req: &lomspb.CreateOrderRequest{
				UserId: 42,
				Items: []*lomspb.Item{
					{Sku: 10, Count: 100},
				},
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					CreateOrder(
						gomock.Any(),
						int64(42),
						[]entity.OrderItem{
							{SKU: 10, Count: 100},
						},
					).
					Return(int64(0), xerrors.ErrInsufficientStock)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.FailedPrecondition,
		},
		{
			name: "unexpected internal error",
			req: &lomspb.CreateOrderRequest{
				UserId: 42,
				Items: []*lomspb.Item{
					{Sku: 10, Count: 1},
				},
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					CreateOrder(
						gomock.Any(),
						int64(42),
						[]entity.OrderItem{
							{SKU: 10, Count: 1},
						},
					).
					Return(int64(0), context.DeadlineExceeded)
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
			lomsService := controllermocks.NewMocklomsService(ctrl)

			test.setupMocks(lomsService)

			server := controllerloms.NewLomsServer(lomsService)

			resp, err := server.CreateOrder(context.Background(), test.req)

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
