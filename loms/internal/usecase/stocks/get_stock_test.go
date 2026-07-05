package stocks_test

import (
	"context"
	"testing"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	stocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/stocks"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/stocks/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStocksServiceGetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sku        uint32
		setupMocks func(stocksRepository *mocksusecase.MockstocksRepository)
		wantCount  uint64
		wantErr    error
	}{
		{
			name:       "invalid input",
			sku:        0,
			setupMocks: func(_ *mocksusecase.MockstocksRepository) {},
			wantCount:  0,
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name: "repository error",
			sku:  404,
			setupMocks: func(stocksRepository *mocksusecase.MockstocksRepository) {
				stocksRepository.EXPECT().
					GetStock(gomock.Any(), uint32(404)).
					Return(uint64(0), xerrors.ErrStockNotFound)
			},
			wantCount: 0,
			wantErr:   xerrors.ErrStockNotFound,
		},
		{
			name: "success",
			sku:  10,
			setupMocks: func(stocksRepository *mocksusecase.MockstocksRepository) {
				stocksRepository.EXPECT().
					GetStock(gomock.Any(), uint32(10)).
					Return(uint64(15), nil)
			},
			wantCount: 15,
			wantErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)

			test.setupMocks(stocksRepository)

			service := stocksusecase.NewStocksService(stocksRepository)

			count, err := service.GetStock(context.Background(), test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantCount, count)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantCount, count)
		})
	}
}
