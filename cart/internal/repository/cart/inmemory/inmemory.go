package inmemory

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

type inMemoryRepository struct {
	mx        sync.RWMutex
	itemsBase map[int64]map[uint32]*ItemRow
}

func NewInMemory() *inMemoryRepository {
	return &inMemoryRepository{itemsBase: make(map[int64]map[uint32]*ItemRow)}
}

func (r *inMemoryRepository) AddItem(_ context.Context, userID int64, item *entity.Item) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.itemsBase[userID]; !ok {
		r.itemsBase[userID] = make(map[uint32]*ItemRow)
	}

	if userItem, ok := r.itemsBase[userID][item.SKU]; ok {
		userItem.Count += item.Count
		return nil
	}

	r.itemsBase[userID][item.SKU] = entityToRow(item)
	return nil
}

func (r *inMemoryRepository) GetItemsByUserID(_ context.Context, userID int64) ([]*entity.Item, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	userItems, ok := r.itemsBase[userID]
	if !ok {
		return nil, nil
	}

	result := make([]*entity.Item, 0, len(userItems))
	for _, row := range userItems {
		item := rowToEntity(row)
		if item != nil {
			result = append(result, item)
		}
	}

	return result, nil
}

func (r *inMemoryRepository) DeleteItemsByUserID(_ context.Context, userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.itemsBase, userID)
	return nil
}

func (r *inMemoryRepository) DeleteItem(_ context.Context, userID int64, sku uint32) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	userItems, ok := r.itemsBase[userID]
	if !ok {
		return nil
	}
	if _, ok := userItems[sku]; !ok {
		return nil
	}
	delete(userItems, sku)

	if len(userItems) == 0 {
		delete(r.itemsBase, userID)
	}
	return nil
}
