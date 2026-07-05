package cart_test

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	cartusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart"
	mocksusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCartServiceCheckoutCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		setupMocks func(
			cartRepository *mocksusecase.MockcartRepository,
			productClient *mocksusecase.MockproductClient,
			lomsClient *mocksusecase.MocklomsClient,
		)
		wantOrderID int64
		wantErr     error
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
			wantOrderID: 0,
			wantErr:     xerrors.ErrInvalidInput,
		},
		{
			name:   "repository get items error",
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
			wantOrderID: 0,
			wantErr:     context.DeadlineExceeded,
		},
		{
			name:   "cart is empty",
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
			wantOrderID: 0,
			wantErr:     xerrors.ErrCartIsEmpty,
		},
		{
			name:   "create order error",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				items := []*entity.Item{
					{SKU: 10, Count: 2},
					{SKU: 12, Count: 1},
				}

				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return(items, nil)

				lomsClient.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(0), xerrors.ErrInsufficientStock)
			},
			wantOrderID: 0,
			wantErr:     xerrors.ErrInsufficientStock,
		},
		{
			name:   "delete items error",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				items := []*entity.Item{
					{SKU: 10, Count: 2},
					{SKU: 13, Count: 1},
				}

				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return(items, nil)

				lomsClient.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(1001), nil)

				cartRepository.EXPECT().
					DeleteItemsByUserID(gomock.Any(), int64(42)).
					Return(context.DeadlineExceeded)
			},
			wantOrderID: 0,
			wantErr:     context.DeadlineExceeded,
		},
		{
			name:   "success",
			userID: 42,
			setupMocks: func(
				cartRepository *mocksusecase.MockcartRepository,
				_ *mocksusecase.MockproductClient,
				lomsClient *mocksusecase.MocklomsClient,
			) {
				items := []*entity.Item{
					{SKU: 10, Count: 2},
					{SKU: 12, Count: 1},
					{SKU: 13, Count: 1},
				}

				cartRepository.EXPECT().
					GetItemsByUserID(gomock.Any(), int64(42)).
					Return(items, nil)

				lomsClient.EXPECT().
					CreateOrder(gomock.Any(), int64(42), items).
					Return(int64(1001), nil)

				cartRepository.EXPECT().
					DeleteItemsByUserID(gomock.Any(), int64(42)).
					Return(nil)
			},
			wantOrderID: 1001,
			wantErr:     nil,
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

			orderID, err := service.CheckoutCart(context.Background(), test.userID)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantOrderID, orderID)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantOrderID, orderID)
		})
	}
}
