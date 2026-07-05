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

func TestStocksServerGetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *stockspb.GetStockRequest
		setupMocks  func(stocksService *controllermocks.MockstocksService)
		wantResp    *stockspb.GetStockResponse
		wantCode    codes.Code
		wantErr     bool
		wantNilResp bool
	}{
		{
			name: "success",
			req: &stockspb.GetStockRequest{
				Sku: 10,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					GetStock(gomock.Any(), uint32(10)).
					Return(uint64(15), nil)
			},
			wantResp: &stockspb.GetStockResponse{
				Count: 15,
			},
			wantErr:     false,
			wantNilResp: false,
		},
		{
			name: "invalid input",
			req: &stockspb.GetStockRequest{
				Sku: 0,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					GetStock(gomock.Any(), uint32(0)).
					Return(uint64(0), xerrors.ErrInvalidInput)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name: "stock not found",
			req: &stockspb.GetStockRequest{
				Sku: 404,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					GetStock(gomock.Any(), uint32(404)).
					Return(uint64(0), xerrors.ErrStockNotFound)
			},
			wantErr:     true,
			wantNilResp: true,
			wantCode:    codes.NotFound,
		},
		{
			name: "unexpected internal error",
			req: &stockspb.GetStockRequest{
				Sku: 12,
			},
			setupMocks: func(stocksService *controllermocks.MockstocksService) {
				stocksService.EXPECT().
					GetStock(gomock.Any(), uint32(12)).
					Return(uint64(0), context.DeadlineExceeded)
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

			resp, err := server.GetStock(context.Background(), test.req)

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
