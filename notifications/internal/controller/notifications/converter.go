package notifications

import notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"

func statusToString(status notificationspb.OrderStatus) string {
	switch status {
	case notificationspb.OrderStatus_ORDER_STATUS_NEW:
		return "new"
	case notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT:
		return "awaiting_payment"
	case notificationspb.OrderStatus_ORDER_STATUS_FAILED:
		return "failed"
	case notificationspb.OrderStatus_ORDER_STATUS_PAID:
		return "paid"
	case notificationspb.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unspecified"
	}
}
