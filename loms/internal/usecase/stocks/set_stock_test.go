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

func TestStocksService_SetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sku        uint32
		count      uint64
		setupMocks func(stocksRepository *mocksusecase.MockstocksRepository)
		wantErr    error
	}{
		{
			name:       "invalid input",
			sku:        0,
			count:      10,
			setupMocks: func(_ *mocksusecase.MockstocksRepository) {},
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name:  "repository error",
			sku:   10,
			count: 25,
			setupMocks: func(stocksRepository *mocksusecase.MockstocksRepository) {
				stocksRepository.EXPECT().
					SetStock(gomock.Any(), uint32(10), uint64(25)).
					Return(context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:  "success",
			sku:   10,
			count: 25,
			setupMocks: func(stocksRepository *mocksusecase.MockstocksRepository) {
				stocksRepository.EXPECT().
					SetStock(gomock.Any(), uint32(10), uint64(25)).
					Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			stocksRepository := mocksusecase.NewMockstocksRepository(ctrl)

			test.setupMocks(stocksRepository)

			service := stocksusecase.NewStocksService(stocksRepository)

			err := service.SetStock(context.Background(), test.sku, test.count)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
