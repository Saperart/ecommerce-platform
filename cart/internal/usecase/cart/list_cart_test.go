package cart_test

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	cartusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCartService_ListCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		setupMocks func(
			cartRepository *mocksusecase.MockcartRepository,
			productClient *mocksusecase.MockproductClient,
			lomsClient *mocksusecase.MocklomsClient,
		)
		wantItems      []*port.Item
		wantTotalPrice uint32
		wantErr        error
	}{
		{
			name:   "invalid input",
			userID: 0,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
			},
			wantItems:      nil,
			wantTotalPrice: 0,
			wantErr:        xerrors.ErrInvalidInput,
		},
		{
			name:   "repository error",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return(nil, context.DeadlineExceeded)
			},
			wantItems:      nil,
			wantTotalPrice: 0,
			wantErr:        context.DeadlineExceeded,
		},
		{
			name:   "empty cart",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return([]*entity.Item{}, nil)
			},
			wantItems:      []*port.Item{},
			wantTotalPrice: 0,
			wantErr:        nil,
		},
		{
			name:   "product client error",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				items := []*entity.Item{
					{SKU: 10, Count: 2},
				}

				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return(items, nil)

				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, xerrors.ErrProductNotFound)
			},
			wantItems:      nil,
			wantTotalPrice: 0,
			wantErr:        xerrors.ErrProductNotFound,
		},
		{
			name:   "success",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				items := []*entity.Item{
					{SKU: 10, Count: 2},
					{SKU: 12, Count: 1},
					{SKU: 13, Count: 1},
				}

				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return(items, nil)

				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(&port.ProductInfo{
						Name:  "Кроссовки",
						Price: 49000,
					}, nil)

				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(12)).
					Return(&port.ProductInfo{
						Name:  "Майка",
						Price: 27000,
					}, nil)

				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(13)).
					Return(&port.ProductInfo{
						Name:  "Стул",
						Price: 11000,
					}, nil)
			},
			wantItems: []*port.Item{
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
			},
			wantTotalPrice: 136000,
			wantErr:        nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			cartRepository := mocksusecase.NewMockcartRepository(ctrl)
			productClient := mocksusecase.NewMockproductClient(ctrl)
			lomsClient := mocksusecase.NewMocklomsClient(ctrl)

			test.setupMocks(cartRepository, productClient, lomsClient)

			service := cartusecase.NewCartService(cartRepository, productClient, lomsClient)

			items, totalPrice, err := service.ListCart(context.Background(), test.userID)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Nil(t, items)
				require.Equal(t, uint32(0), totalPrice)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantItems, items)
			require.Equal(t, test.wantTotalPrice, totalPrice)
		})
	}
}
