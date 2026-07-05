package inmemory

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

type inMemoryRepository struct {
	mx      sync.RWMutex
	nextSKU uint32
	data    map[uint32]*RowProduct
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		data:    make(map[uint32]*RowProduct),
		nextSKU: 1,
	}
}

func (r *inMemoryRepository) GetProduct(_ context.Context, sku uint32) (*entity.Product, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	rowProduct, ok := r.data[sku]
	if !ok {
		return nil, xerrors.ErrProductNotFound
	}
	return rowToProduct(rowProduct), nil
}

func (r *inMemoryRepository) CreateProduct(_ context.Context, name string, price uint32) (uint32, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	sku := r.nextSKU
	r.nextSKU++

	r.data[sku] = &RowProduct{
		SKU:   sku,
		Name:  name,
		Price: price,
	}
	return sku, nil
}
