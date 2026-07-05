package loms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres"
)

func (s *lomsService) enqueueOrderStatusChanged(
	ctx context.Context,
	userID int64,
	orderID int64,
	status entity.OrderStatus,
) error {
	if s.outboxRepository == nil {
		return nil
	}

	notification := port.OrderStatusChangedNotification{
		UserID:  userID,
		OrderID: orderID,
		Status:  toPortOrderStatus(status),
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	return s.outboxRepository.SaveMessage(
		ctx,
		fmt.Sprintf("order-status:%d:%s", orderID, notification.Status),
		outboxrepo.KindNotification,
		payload,
	)
}

func (s *lomsService) OrderStatusChangedNotificationKindHandler(ctx context.Context, payload []byte) error {
	if s.notifications == nil {
		return nil
	}

	var notification port.OrderStatusChangedNotification
	if err := json.Unmarshal(payload, &notification); err != nil {
		return fmt.Errorf("unmarshal notification: %w", err)
	}

	return s.notifications.SendOrderStatusChangedNotification(ctx, notification)
}

func toPortOrderStatus(status entity.OrderStatus) port.OrderStatus {
	switch status {
	case entity.OrderStatusNew:
		return port.OrderStatusNew
	case entity.OrderStatusAwaitingPayment:
		return port.OrderStatusAwaitingPayment
	case entity.OrderStatusFailed:
		return port.OrderStatusFailed
	case entity.OrderStatusPaid:
		return port.OrderStatusPaid
	case entity.OrderStatusCancelled:
		return port.OrderStatusCancelled
	default:
		return port.OrderStatusUnspecified
	}
}
