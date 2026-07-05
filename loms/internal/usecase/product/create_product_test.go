package product_test

import (
	"context"
	"testing"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	productusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/product"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/product/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductServiceCreateProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		nameArg    string
		price      uint32
		setupMocks func(productRepository *mocksusecase.MockproductRepository)
		wantSKU    uint32
		wantErr    error
	}{
		{
			name:       "invalid input empty name",
			nameArg:    "",
			price:      1000,
			setupMocks: func(_ *mocksusecase.MockproductRepository) {},
			wantSKU:    0,
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name:       "invalid input zero price",
			nameArg:    "Майка",
			price:      0,
			setupMocks: func(_ *mocksusecase.MockproductRepository) {},
			wantSKU:    0,
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name:    "repository error",
			nameArg: "Стул",
			price:   11000,
			setupMocks: func(productRepository *mocksusecase.MockproductRepository) {
				productRepository.EXPECT().
					CreateProduct(gomock.Any(), "Стул", uint32(11000)).
					Return(uint32(0), context.DeadlineExceeded)
			},
			wantSKU: 0,
			wantErr: context.DeadlineExceeded,
		},
		{
			name:    "success",
			nameArg: "Майка",
			price:   27000,
			setupMocks: func(productRepository *mocksusecase.MockproductRepository) {
				productRepository.EXPECT().
					CreateProduct(gomock.Any(), "Майка", uint32(27000)).
					Return(uint32(12), nil)
			},
			wantSKU: 12,
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			productRepository := mocksusecase.NewMockproductRepository(ctrl)

			test.setupMocks(productRepository)

			service := productusecase.NewProductService(productRepository)

			sku, err := service.CreateProduct(context.Background(), test.nameArg, test.price)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantSKU, sku)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantSKU, sku)
		})
	}
}
