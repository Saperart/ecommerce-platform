package inmemory

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepositoryGetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sku       uint32
		setupRepo func(repo *inMemoryRepository)
		wantCount uint64
		wantErr   error
	}{
		{
			name: "stock not found",
			sku:  404,
			setupRepo: func(_ *inMemoryRepository) {
			},
			wantCount: 0,
			wantErr:   xerrors.ErrStockNotFound,
		},
		{
			name: "stock exists",
			sku:  10,
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 15)
				require.NoError(t, err)
			},
			wantCount: 15,
			wantErr:   nil,
		},
		{
			name: "zero stock exists",
			sku:  12,
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 12, 0)
				require.NoError(t, err)
			},
			wantCount: 0,
			wantErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			count, err := repo.GetStock(context.Background(), test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantCount, count)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantCount, count)
		})
	}
}

func TestInMemoryRepositorySetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sku       uint32
		count     uint64
		wantCount uint64
	}{
		{
			name:      "set new stock",
			sku:       10,
			count:     15,
			wantCount: 15,
		},
		{
			name:      "overwrite existing stock",
			sku:       12,
			count:     7,
			wantCount: 7,
		},
		{
			name:      "set zero stock",
			sku:       13,
			count:     0,
			wantCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			if test.name == "overwrite existing stock" {
				err := repo.SetStock(context.Background(), test.sku, 100)
				require.NoError(t, err)
			}

			err := repo.SetStock(context.Background(), test.sku, test.count)
			require.NoError(t, err)

			gotCount, err := repo.GetStock(context.Background(), test.sku)
			require.NoError(t, err)
			require.Equal(t, test.wantCount, gotCount)
		})
	}
}

func TestInMemoryRepositoryReleaseStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sku       uint32
		count     uint64
		setupRepo func(repo *inMemoryRepository)
		wantErr   error
		wantCount uint64
	}{
		{
			name:  "stock not found",
			sku:   404,
			count: 5,
			setupRepo: func(_ *inMemoryRepository) {
			},
			wantErr:   xerrors.ErrStockNotFound,
			wantCount: 0,
		},
		{
			name:  "release stock successfully",
			sku:   10,
			count: 5,
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 15)
				require.NoError(t, err)
			},
			wantErr:   nil,
			wantCount: 20,
		},
		{
			name:  "release zero stock",
			sku:   12,
			count: 0,
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 12, 7)
				require.NoError(t, err)
			},
			wantErr:   nil,
			wantCount: 7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			err := repo.ReleaseStock(context.Background(), test.sku, test.count)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)

			gotCount, err := repo.GetStock(context.Background(), test.sku)
			require.NoError(t, err)
			require.Equal(t, test.wantCount, gotCount)
		})
	}
}

func TestInMemoryRepositoryReserveStocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		items           []entity.OrderItem
		setupRepo       func(repo *inMemoryRepository)
		wantErr         error
		wantStocksAfter map[uint32]uint64
	}{
		{
			name: "stock not found",
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
			},
			setupRepo: func(_ *inMemoryRepository) {
			},
			wantErr: xerrors.ErrStockNotFound,
		},
		{
			name: "insufficient stock",
			items: []entity.OrderItem{
				{SKU: 10, Count: 20},
			},
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 15)
				require.NoError(t, err)
			},
			wantErr: xerrors.ErrInsufficientStock,
		},
		{
			name: "reserve one item successfully",
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
			},
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 15)
				require.NoError(t, err)
			},
			wantErr: nil,
			wantStocksAfter: map[uint32]uint64{
				10: 13,
			},
		},
		{
			name: "reserve multiple items successfully",
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 12, Count: 1},
				{SKU: 13, Count: 3},
			},
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 15)
				require.NoError(t, err)
				err = repo.SetStock(context.Background(), 12, 7)
				require.NoError(t, err)
				err = repo.SetStock(context.Background(), 13, 10)
				require.NoError(t, err)
			},
			wantErr: nil,
			wantStocksAfter: map[uint32]uint64{
				10: 13,
				12: 6,
				13: 7,
			},
		},
		{
			name: "reserve is atomic on failure",
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 12, Count: 100},
			},
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 15)
				require.NoError(t, err)
				err = repo.SetStock(context.Background(), 12, 7)
				require.NoError(t, err)
			},
			wantErr: xerrors.ErrInsufficientStock,
			wantStocksAfter: map[uint32]uint64{
				10: 15,
				12: 7,
			},
		},
		{
			name: "reserve repeated same sku",
			items: []entity.OrderItem{
				{SKU: 10, Count: 2},
				{SKU: 10, Count: 3},
			},
			setupRepo: func(repo *inMemoryRepository) {
				err := repo.SetStock(context.Background(), 10, 10)
				require.NoError(t, err)
			},
			wantErr: nil,
			wantStocksAfter: map[uint32]uint64{
				10: 5,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			err := repo.ReserveStocks(context.Background(), test.items)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}

			for sku, wantCount := range test.wantStocksAfter {
				gotCount, getErr := repo.GetStock(context.Background(), sku)
				require.NoError(t, getErr)
				require.Equal(t, wantCount, gotCount)
			}
		})
	}
}
