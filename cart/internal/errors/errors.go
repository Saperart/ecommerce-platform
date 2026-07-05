package errors

import "errors"

var (
	ErrItemNotFound      = errors.New("item not found")
	ErrCartNotFound      = errors.New("cart not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrStockNotFound     = errors.New("stock not found")
	ErrCartIsEmpty       = errors.New("cart is empty")
)
