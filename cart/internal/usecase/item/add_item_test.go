package item_test

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	itemusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestItemService_AddItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		sku        uint32
		count      uint32
		setupMocks func(
			cartRepository *mocksusecase.MockcartRepository,
			productClient *mocksusecase.MockproductClient,
			lomsClient *mocksusecase.MocklomsClient,
		)
		wantErr error
	}{
		{
			name:   "invalid input user id",
			userID: 0,
			sku:    10,
			count:  1,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
			},
			wantErr: xerrors.ErrInvalidInput,
		},
		{
			name:   "invalid input sku",
			userID: 42,
			sku:    0,
			count:  1,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
			},
			wantErr: xerrors.ErrInvalidInput,
		},
		{
			name:   "invalid input count",
			userID: 42,
			sku:    10,
			count:  0,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
			},
			wantErr: xerrors.ErrInvalidInput,
		},
		{
			name:   "product not found",
			userID: 42,
			sku:    10,
			count:  1,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, xerrors.ErrProductNotFound)
			},
			wantErr: xerrors.ErrProductNotFound,
		},
		{
			name:   "unexpected product client error",
			userID: 42,
			sku:    10,
			count:  1,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:   "stock client error",
			userID: 42,
			sku:    10,
			count:  1,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, nil)

				lomsClient.EXPECT().
					GetStock(gomock.Any(), uint32(10)).
					Return(uint64(0), context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:   "insufficient stock",
			userID: 42,
			sku:    10,
			count:  5,
			setupMocks: func(
				_ *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, nil)

				lomsClient.EXPECT().
					GetStock(gomock.Any(), uint32(10)).
					Return(uint64(3), nil)
			},
			wantErr: xerrors.ErrInsufficientStock,
		},
		{
			name:   "repository add item error",
			userID: 42,
			sku:    10,
			count:  2,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, nil)

				lomsClient.EXPECT().
					GetStock(gomock.Any(), uint32(10)).
					Return(uint64(10), nil)

				cartRepository.EXPECT().
					AddItem(gomock.Any(), int64(42), &entity.Item{
						SKU:   10,
						Count: 2,
					}).
					Return(context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:   "success",
			userID: 42,
			sku:    10,
			count:  2,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				productClient *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				productClient.EXPECT().
					GetProduct(gomock.Any(), uint32(10)).
					Return(nil, nil)

				lomsClient.EXPECT().
					GetStock(gomock.Any(), uint32(10)).
					Return(uint64(10), nil)

				cartRepository.EXPECT().
					AddItem(gomock.Any(), int64(42), &entity.Item{
						SKU:   10,
						Count: 2,
					}).
					Return(nil)
			},
			wantErr: nil,
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

			service := itemusecase.NewItemService(cartRepository, productClient, lomsClient)

			err := service.AddItem(context.Background(), test.userID, test.sku, test.count)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
