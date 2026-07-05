package inmemory

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepositoryCreateOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     int64
		items      []entity.OrderItem
		wantUserID int64
		wantStatus entity.OrderStatus
		wantItems  []entity.OrderItem
	}{
		{
			name:   "create order with one item",
			userID: 42,
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
			},
			wantUserID: 42,
			wantStatus: entity.OrderStatusAwaitingPayment,
			wantItems: []entity.OrderItem{
				{SKU: 10, Count: 2},
			},
		},
		{
			name:   "create order with multiple items",
			userID: 100,
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 12, Count: 1},
				{SKU: 13, Count: 3},
			},
			wantUserID: 100,
			wantStatus: entity.OrderStatusAwaitingPayment,
			wantItems: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 12, Count: 1},
				{SKU: 13, Count: 3},
			},
		},
		{
			name:       "create order with empty items",
			userID:     7,
			items:      []entity.OrderItem{},
			wantUserID: 7,
			wantStatus: entity.OrderStatusAwaitingPayment,
			wantItems:  []entity.OrderItem{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			orderID, err := repo.CreateOrder(context.Background(), test.userID, test.items)
			require.NoError(t, err)
			require.Equal(t, int64(1), orderID)

			order, err := repo.GetOrder(context.Background(), orderID)
			require.NoError(t, err)

			require.Equal(t, orderID, order.ID)
			require.Equal(t, test.wantUserID, order.UserID)
			require.Equal(t, test.wantStatus, order.Status)
			require.Equal(t, test.wantItems, order.Items)
			require.False(t, order.CreatedAt.IsZero())
			require.False(t, order.UpdatedAt.IsZero())
		})
	}
}

func TestInMemoryRepositoryCreateOrderIncrementsID(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()

	firstID, err := repo.CreateOrder(context.Background(), 42, []entity.OrderItem{
		{SKU: 10, Count: 1},
	})
	require.NoError(t, err)

	secondID, err := repo.CreateOrder(context.Background(), 42, []entity.OrderItem{
		{SKU: 12, Count: 2},
	})
	require.NoError(t, err)

	require.Equal(t, int64(1), firstID)
	require.Equal(t, int64(2), secondID)
}

func TestInMemoryRepositoryCreateOrderCopiesItems(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()

	items := []entity.OrderItem{
		{SKU: 10, Count: 2},
		{SKU: 12, Count: 1},
	}

	orderID, err := repo.CreateOrder(context.Background(), 42, items)
	require.NoError(t, err)

	items[0].Count = 999

	order, err := repo.GetOrder(context.Background(), orderID)
	require.NoError(t, err)

	require.Equal(t, []entity.OrderItem{
		{SKU: 10, Count: 2},
		{SKU: 12, Count: 1},
	}, order.Items)
}

func TestInMemoryRepositoryGetOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		orderID   int64
		setupRepo func(repo *inMemoryRepository)
		wantErr   error
	}{
		{
			name:    "order not found",
			orderID: 404,
			setupRepo: func(_ *inMemoryRepository) {
			},
			wantErr: xerrors.ErrOrderNotFound,
		},
		{
			name:    "order exists",
			orderID: 1,
			setupRepo: func(repo *inMemoryRepository) {
				_, err := repo.CreateOrder(context.Background(), 42, []entity.OrderItem{
					{SKU: 10, Count: 2},
				})
				require.NoError(t, err)
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			order, err := repo.GetOrder(context.Background(), test.orderID)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Nil(t, order)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			require.Equal(t, test.orderID, order.ID)
		})
	}
}

func TestInMemoryRepositoryGetOrderReturnsCopy(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()

	orderID, err := repo.CreateOrder(context.Background(), 42, []entity.OrderItem{
		{SKU: 10, Count: 2},
	})
	require.NoError(t, err)

	order, err := repo.GetOrder(context.Background(), orderID)
	require.NoError(t, err)

	order.Status = entity.OrderStatusPaid
	order.Items[0].Count = 999

	freshOrder, err := repo.GetOrder(context.Background(), orderID)
	require.NoError(t, err)

	require.Equal(t, entity.OrderStatusAwaitingPayment, freshOrder.Status)
	require.Equal(t, []entity.OrderItem{
		{SKU: 10, Count: 2},
	}, freshOrder.Items)
}

func TestInMemoryRepositoryDeleteOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		orderID   int64
		setupRepo func(repo *inMemoryRepository)
		wantErr   error
	}{
		{
			name:    "delete existing order",
			orderID: 1,
			setupRepo: func(repo *inMemoryRepository) {
				_, err := repo.CreateOrder(context.Background(), 42, []entity.OrderItem{
					{SKU: 10, Count: 2},
				})
				require.NoError(t, err)
			},
			wantErr: nil,
		},
		{
			name:    "delete non existing order is no-op",
			orderID: 404,
			setupRepo: func(_ *inMemoryRepository) {
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			err := repo.DeleteOrder(context.Background(), test.orderID)
			require.NoError(t, err)

			order, getErr := repo.GetOrder(context.Background(), test.orderID)
			require.Error(t, getErr)
			require.ErrorIs(t, getErr, xerrors.ErrOrderNotFound)
			require.Nil(t, order)
		})
	}
}

func TestInMemoryRepositorySetOrderStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		orderID    int64
		newStatus  entity.OrderStatus
		setupRepo  func(repo *inMemoryRepository)
		wantErr    error
		wantStatus entity.OrderStatus
	}{
		{
			name:      "order not found",
			orderID:   404,
			newStatus: entity.OrderStatusPaid,
			setupRepo: func(_ *inMemoryRepository) {
			},
			wantErr: xerrors.ErrOrderNotFound,
		},
		{
			name:      "set status successfully",
			orderID:   1,
			newStatus: entity.OrderStatusCancelled,
			setupRepo: func(repo *inMemoryRepository) {
				_, err := repo.CreateOrder(context.Background(), 42, []entity.OrderItem{
					{SKU: 10, Count: 2},
				})
				require.NoError(t, err)
			},
			wantErr:    nil,
			wantStatus: entity.OrderStatusCancelled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			var beforeUpdatedAt int64
			if test.wantErr == nil {
				order, err := repo.GetOrder(context.Background(), test.orderID)
				require.NoError(t, err)
				beforeUpdatedAt = order.UpdatedAt.UnixNano()
			}

			err := repo.SetOrderStatus(context.Background(), test.orderID, test.newStatus)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)

			order, err := repo.GetOrder(context.Background(), test.orderID)
			require.NoError(t, err)
			require.Equal(t, test.wantStatus, order.Status)
			require.GreaterOrEqual(t, order.UpdatedAt.UnixNano(), beforeUpdatedAt)
		})
	}
}
