package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
)

type inMemoryRepository struct {
	mx     sync.RWMutex
	nextID int64
	orders map[int64]*entity.Order
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		nextID: 1,
		orders: make(map[int64]*entity.Order),
	}
}

func (r *inMemoryRepository) CreateOrder(_ context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	id := r.nextID
	r.nextID++
	now := time.Now().UTC()

	orderItems := make([]entity.OrderItem, len(items))
	copy(orderItems, items)

	order := &entity.Order{
		ID:        id,
		UserID:    userID,
		Status:    entity.OrderStatusAwaitingPayment,
		Items:     orderItems,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.orders[id] = order
	return id, nil
}

func (r *inMemoryRepository) DeleteOrder(_ context.Context, orderID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	delete(r.orders, orderID)
	return nil
}

func (r *inMemoryRepository) GetOrder(_ context.Context, orderID int64) (*entity.Order, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	order, ok := r.orders[orderID]
	if !ok {
		return nil, xerrors.ErrOrderNotFound
	}
	orderCopy := *order

	itemsCopy := make([]entity.OrderItem, len(order.Items))
	copy(itemsCopy, order.Items)
	orderCopy.Items = itemsCopy
	return &orderCopy, nil
}

func (r *inMemoryRepository) SetOrderStatus(_ context.Context, orderID int64, status entity.OrderStatus) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	order, ok := r.orders[orderID]
	if !ok {
		return xerrors.ErrOrderNotFound
	}
	if (status == entity.OrderStatusPaid || status == entity.OrderStatusCancelled) &&
		order.Status != entity.OrderStatusAwaitingPayment {
		return xerrors.ErrInvalidOrderStatus
	}

	order.Status = status
	order.UpdatedAt = time.Now().UTC()
	return nil
}
