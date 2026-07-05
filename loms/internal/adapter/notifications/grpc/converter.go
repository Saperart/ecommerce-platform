package grpc

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
)

func toProtoStatus(status port.OrderStatus) notificationspb.OrderStatus {
	switch status {
	case port.OrderStatusNew:
		return notificationspb.OrderStatus_ORDER_STATUS_NEW
	case port.OrderStatusAwaitingPayment:
		return notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case port.OrderStatusFailed:
		return notificationspb.OrderStatus_ORDER_STATUS_FAILED
	case port.OrderStatusPaid:
		return notificationspb.OrderStatus_ORDER_STATUS_PAID
	case port.OrderStatusCancelled:
		return notificationspb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return notificationspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
