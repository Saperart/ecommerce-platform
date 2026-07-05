package loms_test

import (
	"context"
	"testing"

	controllerloms "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLOMSServerPayOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *lomspb.PayOrderRequest
		setupMocks  func(lomsService *controllermocks.MocklomsService)
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &lomspb.PayOrderRequest{
				OrderId: 1001,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					PayOrder(gomock.Any(), int64(1001)).
					Return(nil)
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &lomspb.PayOrderRequest{
				OrderId: 0,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					PayOrder(gomock.Any(), int64(0)).
					Return(xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "invalid order status",
			req: &lomspb.PayOrderRequest{
				OrderId: 1002,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					PayOrder(gomock.Any(), int64(1002)).
					Return(xerrors.ErrInvalidOrderStatus)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.FailedPrecondition,
		},
		{
			name: "order not found",
			req: &lomspb.PayOrderRequest{
				OrderId: 404,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					PayOrder(gomock.Any(), int64(404)).
					Return(xerrors.ErrOrderNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "unexpected internal error",
			req: &lomspb.PayOrderRequest{
				OrderId: 1003,
			},
			setupMocks: func(lomsService *controllermocks.MocklomsService) {
				lomsService.EXPECT().
					PayOrder(gomock.Any(), int64(1003)).
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
			lomsService := controllermocks.NewMocklomsService(ctrl)

			test.setupMocks(lomsService)

			server := controllerloms.NewLomsServer(lomsService)

			resp, err := server.PayOrder(context.Background(), test.req)

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
