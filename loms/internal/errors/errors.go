package errors

import "errors"

var (
	ErrStockNotFound      = errors.New("stock not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrProductNotFound    = errors.New("product not found")
	ErrOrderNotFound      = errors.New("order not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInvalidOrderStatus = errors.New("invalid order status")
)
