package controller

import notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"

type API struct {
	Notifications notificationspb.NotificationsServer
}

func New(notifications notificationspb.NotificationsServer) *API {
	return &API{Notifications: notifications}
}
