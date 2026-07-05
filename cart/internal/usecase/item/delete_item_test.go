package item_test

import (
	"context"
	"testing"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	itemusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item"
	usecasemocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestItemService_DeleteItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		sku        uint32
		setupMocks func(
			cartRepository *usecasemocks.MockcartRepository,
			productClient *usecasemocks.MockproductClient,
			lomsClient *usecasemocks.MocklomsClient,
		)
		wantErr error
	}{
		{
			name:   "invalid input user id",
			userID: 0,
			sku:    10,
			setupMocks: func(
				_ *usecasemocks.MockcartRepository,
				_ *usecasemocks.MockproductClient,
				_ *usecasemocks.MocklomsClient,
			) {
			},
			wantErr: xerrors.ErrInvalidInput,
		},
		{
			name:   "invalid input sku",
			userID: 42,
			sku:    0,
			setupMocks: func(
				_ *usecasemocks.MockcartRepository,
				_ *usecasemocks.MockproductClient,
				_ *usecasemocks.MocklomsClient,
			) {
			},
			wantErr: xerrors.ErrInvalidInput,
		},
		{
			name:   "repository returns cart not found",
			userID: 42,
			sku:    10,
			setupMocks: func(
				cartRepository *usecasemocks.MockcartRepository,
				_ *usecasemocks.MockproductClient,
				_ *usecasemocks.MocklomsClient,
			) {
				cartRepository.EXPECT().
					DeleteItem(gomock.Any(), int64(42), uint32(10)).
					Return(xerrors.ErrCartNotFound)
			},
			wantErr: xerrors.ErrCartNotFound,
		},
		{
			name:   "repository returns item not found",
			userID: 42,
			sku:    10,
			setupMocks: func(
				cartRepository *usecasemocks.MockcartRepository,
				_ *usecasemocks.MockproductClient,
				_ *usecasemocks.MocklomsClient,
			) {
				cartRepository.EXPECT().
					DeleteItem(gomock.Any(), int64(42), uint32(10)).
					Return(xerrors.ErrItemNotFound)
			},
			wantErr: xerrors.ErrItemNotFound,
		},
		{
			name:   "repository unexpected error",
			userID: 42,
			sku:    10,
			setupMocks: func(
				cartRepository *usecasemocks.MockcartRepository,
				_ *usecasemocks.MockproductClient,
				_ *usecasemocks.MocklomsClient,
			) {
				cartRepository.EXPECT().
					DeleteItem(gomock.Any(), int64(42), uint32(10)).
					Return(context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:   "success",
			userID: 42,
			sku:    10,
			setupMocks: func(
				cartRepository *usecasemocks.MockcartRepository,
				_ *usecasemocks.MockproductClient,
				_ *usecasemocks.MocklomsClient,
			) {
				cartRepository.EXPECT().
					DeleteItem(gomock.Any(), int64(42), uint32(10)).
					Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			cartRepository := usecasemocks.NewMockcartRepository(ctrl)
			productClient := usecasemocks.NewMockproductClient(ctrl)
			lomsClient := usecasemocks.NewMocklomsClient(ctrl)

			test.setupMocks(cartRepository, productClient, lomsClient)

			service := itemusecase.NewItemService(cartRepository, productClient, lomsClient)

			err := service.DeleteItem(context.Background(), test.userID, test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
