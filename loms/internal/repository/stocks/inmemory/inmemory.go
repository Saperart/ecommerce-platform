package inmemory

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

// в данном случае можно было завести entity.Stock и model.StockRow,
// но я думаю что это будет просто усложнение логики
type inMemoryRepository struct {
	mx   sync.RWMutex
	data map[uint32]uint64
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{data: make(map[uint32]uint64)}
}

func (r *inMemoryRepository) GetStock(_ context.Context, sku uint32) (uint64, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	count, ok := r.data[sku]
	if !ok {
		return 0, xerrors.ErrStockNotFound
	}
	return count, nil
}

func (r *inMemoryRepository) ReserveStocks(_ context.Context, items []entity.OrderItem) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, item := range items {
		cnt, ok := r.data[item.SKU]
		if !ok {
			return xerrors.ErrStockNotFound
		}
		if cnt < uint64(item.Count) {
			return xerrors.ErrInsufficientStock
		}
	}

	for _, item := range items {
		r.data[item.SKU] -= uint64(item.Count)
	}
	return nil
}

func (r *inMemoryRepository) ReleaseStock(_ context.Context, sku uint32, count uint64) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	_, ok := r.data[sku]
	if !ok {
		// мы возвращаем количество товара, а значит у нас уже в мап есть этот sku,
		// поэтому скоре всего эта ошибка никогда не произойдет
		return xerrors.ErrStockNotFound
	}
	r.data[sku] += count
	return nil
}

func (r *inMemoryRepository) SetStock(_ context.Context, sku uint32, count uint64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.data[sku] = count
	return nil
}
