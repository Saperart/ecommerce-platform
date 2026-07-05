package inmemory

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepositoryCreateProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		nameArg   string
		price     uint32
		wantSKU   uint32
		wantName  string
		wantPrice uint32
	}{
		{
			name:      "create first product",
			nameArg:   "Кроссовки",
			price:     49000,
			wantSKU:   1,
			wantName:  "Кроссовки",
			wantPrice: 49000,
		},
		{
			name:      "create product with another name",
			nameArg:   "Майка",
			price:     27000,
			wantSKU:   1,
			wantName:  "Майка",
			wantPrice: 27000,
		},
		{
			name:      "create product with zero price",
			nameArg:   "Стул",
			price:     0,
			wantSKU:   1,
			wantName:  "Стул",
			wantPrice: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()

			sku, err := repo.CreateProduct(context.Background(), test.nameArg, test.price)
			require.NoError(t, err)
			require.Equal(t, test.wantSKU, sku)

			product, err := repo.GetProduct(context.Background(), sku)
			require.NoError(t, err)
			require.Equal(t, &entity.Product{
				SKU:   test.wantSKU,
				Name:  test.wantName,
				Price: test.wantPrice,
			}, product)
		})
	}
}

func TestInMemoryRepositoryCreateProductIncrementsSKU(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()

	firstSKU, err := repo.CreateProduct(context.Background(), "Кроссовки", 49000)
	require.NoError(t, err)

	secondSKU, err := repo.CreateProduct(context.Background(), "Майка", 27000)
	require.NoError(t, err)

	thirdSKU, err := repo.CreateProduct(context.Background(), "Стул", 11000)
	require.NoError(t, err)

	require.Equal(t, uint32(1), firstSKU)
	require.Equal(t, uint32(2), secondSKU)
	require.Equal(t, uint32(3), thirdSKU)
}

func TestInMemoryRepositoryGetProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sku       uint32
		setupRepo func(repo *inMemoryRepository)
		want      *entity.Product
		wantErr   error
	}{
		{
			name: "product not found",
			sku:  404,
			setupRepo: func(_ *inMemoryRepository) {
			},
			want:    nil,
			wantErr: xerrors.ErrProductNotFound,
		},
		{
			name: "product exists",
			sku:  1,
			setupRepo: func(repo *inMemoryRepository) {
				_, err := repo.CreateProduct(context.Background(), "Кроссовки", 49000)
				require.NoError(t, err)
			},
			want: &entity.Product{
				SKU:   1,
				Name:  "Кроссовки",
				Price: 49000,
			},
			wantErr: nil,
		},
		{
			name: "get second product",
			sku:  2,
			setupRepo: func(repo *inMemoryRepository) {
				_, err := repo.CreateProduct(context.Background(), "Кроссовки", 49000)
				require.NoError(t, err)
				_, err = repo.CreateProduct(context.Background(), "Майка", 27000)
				require.NoError(t, err)
			},
			want: &entity.Product{
				SKU:   2,
				Name:  "Майка",
				Price: 27000,
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			test.setupRepo(repo)

			product, err := repo.GetProduct(context.Background(), test.sku)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Nil(t, product)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, product)
		})
	}
}

func TestInMemoryRepositoryGetProductReturnsCopy(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()

	sku, err := repo.CreateProduct(context.Background(), "Кроссовки", 49000)
	require.NoError(t, err)

	product, err := repo.GetProduct(context.Background(), sku)
	require.NoError(t, err)

	product.Name = "Испорчено"
	product.Price = 1

	freshProduct, err := repo.GetProduct(context.Background(), sku)
	require.NoError(t, err)

	require.Equal(t, &entity.Product{
		SKU:   sku,
		Name:  "Кроссовки",
		Price: 49000,
	}, freshProduct)
}
