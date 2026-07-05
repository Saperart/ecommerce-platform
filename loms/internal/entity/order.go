package entity

import "time"

type OrderStatus int32

const (
	OrderStatusUnspecified     OrderStatus = 0
	OrderStatusNew             OrderStatus = 1
	OrderStatusAwaitingPayment OrderStatus = 2
	OrderStatusFailed          OrderStatus = 3
	OrderStatusPaid            OrderStatus = 4
	OrderStatusCancelled       OrderStatus = 5
)

type OrderItem struct {
	SKU   uint32
	Count uint32
}

type Order struct {
	ID        int64
	UserID    int64
	Status    OrderStatus
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}
