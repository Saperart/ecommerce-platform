package grpc

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
)

//go:generate mockgen -destination=mocks/notifications_client_mock.go -package=mocks github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1 NotificationsClient

type notificationsClient struct {
	client notificationspb.NotificationsClient
}

func NewNotificationsClient(client notificationspb.NotificationsClient) *notificationsClient {
	return &notificationsClient{client: client}
}

func (c *notificationsClient) SendOrderStatusChangedNotification(
	ctx context.Context,
	notification port.OrderStatusChangedNotification,
) error {
	_, err := c.client.SendOrderStatusChangedNotification(ctx, &notificationspb.OrderStatusChangedNotificationRequest{
		UserId:  notification.UserID,
		OrderId: notification.OrderID,
		Status:  toProtoStatus(notification.Status),
	})
	return err
}
