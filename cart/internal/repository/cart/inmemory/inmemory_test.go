package inmemory

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/stretchr/testify/require"
)

func seedRepoItems(t *testing.T, repo *inMemoryRepository, userID int64, items []*entity.Item) {
	t.Helper()

	for _, item := range items {
		err := repo.AddItem(context.Background(), userID, item)
		require.NoError(t, err)
	}
}

func requireCartItems(t *testing.T, repo *inMemoryRepository, userID int64, want []*entity.Item) {
	t.Helper()

	gotItems, err := repo.GetItemsByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.ElementsMatch(t, want, gotItems)
}

func TestInMemoryRepositoryAddItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		userID       int64
		initialItems []*entity.Item
		itemToAdd    *entity.Item
		wantItems    []*entity.Item
	}{
		{
			name:         "add item to empty cart",
			userID:       42,
			initialItems: nil,
			itemToAdd: &entity.Item{
				SKU:   10,
				Count: 2,
			},
			wantItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
			},
		},
		{
			name:   "add new item to existing cart",
			userID: 42,
			initialItems: []*entity.Item{
				{
					SKU:   12,
					Count: 1,
				},
			},
			itemToAdd: &entity.Item{
				SKU:   10,
				Count: 2,
			},
			wantItems: []*entity.Item{
				{
					SKU:   12,
					Count: 1,
				},
				{
					SKU:   10,
					Count: 2,
				},
			},
		},
		{
			name:   "merge counts for same sku",
			userID: 42,
			initialItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
			},
			itemToAdd: &entity.Item{
				SKU:   10,
				Count: 3,
			},
			wantItems: []*entity.Item{
				{
					SKU:   10,
					Count: 5,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemory()
			seedRepoItems(t, repo, test.userID, test.initialItems)

			err := repo.AddItem(context.Background(), test.userID, test.itemToAdd)
			require.NoError(t, err)

			requireCartItems(t, repo, test.userID, test.wantItems)
		})
	}
}

func TestInMemoryRepositoryGetItemsByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		userID       int64
		initialItems []*entity.Item
		wantItems    []*entity.Item
	}{
		{
			name:         "cart does not exist",
			userID:       42,
			initialItems: nil,
			wantItems:    nil,
		},
		{
			name:   "cart has items",
			userID: 42,
			initialItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
				{
					SKU:   12,
					Count: 1,
				},
			},
			wantItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
				{
					SKU:   12,
					Count: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemory()
			seedRepoItems(t, repo, test.userID, test.initialItems)

			requireCartItems(t, repo, test.userID, test.wantItems)
		})
	}
}

func TestInMemoryRepositoryDeleteItemsByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		userID            int64
		initialItems      []*entity.Item
		wantItemsAfterDel []*entity.Item
	}{
		{
			name:              "delete from empty cart",
			userID:            42,
			initialItems:      nil,
			wantItemsAfterDel: nil,
		},
		{
			name:   "delete existing cart",
			userID: 42,
			initialItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
				{
					SKU:   12,
					Count: 1,
				},
			},
			wantItemsAfterDel: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemory()
			seedRepoItems(t, repo, test.userID, test.initialItems)

			err := repo.DeleteItemsByUserID(context.Background(), test.userID)
			require.NoError(t, err)

			requireCartItems(t, repo, test.userID, test.wantItemsAfterDel)
		})
	}
}

func TestInMemoryRepositoryDeleteItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		userID            int64
		sku               uint32
		initialItems      []*entity.Item
		wantErr           error
		wantItemsAfterDel []*entity.Item
	}{
		{
			name:              "cart not found",
			userID:            42,
			sku:               10,
			initialItems:      nil,
			wantErr:           nil,
			wantItemsAfterDel: nil,
		},
		{
			name:   "item not found",
			userID: 42,
			sku:    99,
			initialItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
			},
			wantErr: nil,
			wantItemsAfterDel: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
			},
		},
		{
			name:   "delete one item from multi item cart",
			userID: 42,
			sku:    10,
			initialItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
				{
					SKU:   12,
					Count: 1,
				},
			},
			wantErr: nil,
			wantItemsAfterDel: []*entity.Item{
				{
					SKU:   12,
					Count: 1,
				},
			},
		},
		{
			name:   "delete last item removes cart",
			userID: 42,
			sku:    10,
			initialItems: []*entity.Item{
				{
					SKU:   10,
					Count: 2,
				},
			},
			wantErr:           nil,
			wantItemsAfterDel: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemory()
			seedRepoItems(t, repo, test.userID, test.initialItems)

			err := repo.DeleteItem(context.Background(), test.userID, test.sku)
			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}

			requireCartItems(t, repo, test.userID, test.wantItemsAfterDel)
		})
	}
}
