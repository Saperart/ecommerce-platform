package product_test

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	productusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/product"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/product/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductServiceGetProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sku        uint32
		setupMocks func(productRepository *mocksusecase.MockproductRepository)
		wantResult *entity.Product
		wantErr    error
	}{
		{
			name:       "invalid input",
			sku:        0,
			setupMocks: func(_ *mocksusecase.MockproductRepository) {},
			wantResult: nil,
			wantErr:    xerrors.ErrInvalidInput,
		},
		{
			name: "repository error",
			sku:  404,
			setupMocks: func(productRepository *mocksusecase.MockproductRepository) {
				productRepository.EXPECT().
					GetProduct(gomock.Any(), uint32(404)).
					Return(nil, xerrors.ErrProductNotFound)
			},
			wantResult: nil,
			wantErr:    xerrors.ErrProductNotFound,
		},
		{
			name: "success",
			sku:  10,
			setupMocks: func(productRepository *mocksusecase.MockproductRepository) {
				productRepository.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(&entity.Product{
						SKU:   10,
						Name:  "Кроссовки",
						Price: 49000,
					}, nil)
			},
			wantResult: &entity.Product{
				SKU:   10,
				Name:  "Кроссовки",
				Price: 49000,
			},
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

			gotResult, err := service.GetProduct(context.Background(), test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Nil(t, gotResult)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}
