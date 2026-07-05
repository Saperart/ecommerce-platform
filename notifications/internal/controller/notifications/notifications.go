package notifications

import (
	"net/http"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"go.uber.org/zap"
)

type notificationsServer struct {
	notificationspb.UnimplementedNotificationsServer

	logger *zap.Logger
	cfg    *config.Config
	client *http.Client
}

type callbackPayload struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

func NewNotificationsServer(logger *zap.Logger, cfg *config.Config) *notificationsServer {
	return &notificationsServer{
		logger: logger,
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Callback.Timeout},
	}
}
