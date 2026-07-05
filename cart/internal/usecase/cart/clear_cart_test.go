package cart_test

import (
	"context"
	"testing"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	cartusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCartService_ClearCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		setupMocks func(
			cartRepository *mocksusecase.MockcartRepository,
			productClient *mocksusecase.MockproductClient,
			lomsClient *mocksusecase.MocklomsClient,
		)
		wantErr error
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
			wantErr: xerrors.ErrInvalidInput,
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
					DeleteItemsByUserID(gomock.Any(), int64(42)).
					Return(context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:   "success",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				_ *mocksusecase.MocklomsClient,
			) {
				cartRepository.EXPECT().
					DeleteItemsByUserID(gomock.Any(), int64(42)).
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

			service := cartusecase.NewCartService(cartRepository, productClient, lomsClient)

			err := service.ClearCart(context.Background(), test.userID)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
