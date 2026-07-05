package loms_test

import (
	"context"
	"testing"
	"time"

	controllerloms "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLOMSServerGetOrder(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		req         *lomspb.GetOrderRequest
		setupMocks  func(lomsService *controllermocks.MocklomsService)
		wantResp    *lomspb.GetOrderResponse
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &lomspb.GetOrderRequest{
				OrderId: 1001,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					GetOrder(gomock.Any(), int64(1001)).
					Return(&entity.Order{
						ID:     1001,
						UserID: 42,
						Status: entity.OrderStatusAwaitingPayment,
						Items: []entity.OrderItem{
							{SKU: 10, Count: 2},
							{SKU: 12, Count: 1},
						},
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					}, nil)
			},
			wantResp: &lomspb.GetOrderResponse{
				Status: lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT,
				UserId: 42,
				Items: []*lomspb.Item{
					{Sku: 10, Count: 2},
					{Sku: 12, Count: 1},
				},
				CreatedAt: timestamppb.New(createdAt),
				UpdatedAt: timestamppb.New(updatedAt),
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &lomspb.GetOrderRequest{
				OrderId: 0,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					GetOrder(gomock.Any(), int64(0)).
					Return(nil, xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "order not found",
			req: &lomspb.GetOrderRequest{
				OrderId: 404,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					GetOrder(gomock.Any(), int64(404)).
					Return(nil, xerrors.ErrOrderNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "unexpected internal error",
			req: &lomspb.GetOrderRequest{
				OrderId: 1002,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					GetOrder(gomock.Any(), int64(1002)).
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
			lomsService := controllermocks.NewMocklomsService(ctrl)

			test.setupMocks(lomsService)

			server := controllerloms.NewLomsServer(lomsService)

			resp, err := server.GetOrder(context.Background(), test.req)

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
