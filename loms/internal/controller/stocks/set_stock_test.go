package stocks_test

import (
	"context"
	"testing"

	controllerstocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
	controllermocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks/mocks"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStocksServerSetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *stockspb.SetStockRequest
		setupMocks  func(stocksService *controllermocks.MockstocksService)
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &stockspb.SetStockRequest{
				Sku:   10,
				Count: 25,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					SetStock(gomock.Any(), uint32(10), uint64(25)).
					Return(nil)
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &stockspb.SetStockRequest{
				Sku:   0,
				Count: 25,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					SetStock(gomock.Any(), uint32(0), uint64(25)).
					Return(xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "unexpected internal error",
			req: &stockspb.SetStockRequest{
				Sku:   13,
				Count: 5,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					SetStock(gomock.Any(), uint32(13), uint64(5)).
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
			stocksService := controllermocks.NewMockstocksService(ctrl)

			test.setupMocks(stocksService)

			server := controllerstocks.NewStocksServer(stocksService)

			resp, err := server.SetStock(context.Background(), test.req)

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
