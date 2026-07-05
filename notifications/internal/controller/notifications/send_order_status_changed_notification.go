package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *notificationsServer) SendOrderStatusChangedNotification(
	ctx context.Context,
	req *notificationspb.OrderStatusChangedNotificationRequest,
) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	callbackURL := s.cfg.CallbackURL()
	if callbackURL == "" {
		return &emptypb.Empty{}, nil
	}

	body, err := json.Marshal(callbackPayload{
		UserID:  req.GetUserId(),
		OrderID: req.GetOrderId(),
		Status:  statusToString(req.GetStatus()),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "marshal callback payload")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
	if err != nil {
		return nil, status.Error(codes.Internal, "create callback request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "send callback: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("callback status: %d", resp.StatusCode))
	}

	s.logger.Info(
		"order status notification sent",
		zap.Int64("user_id", req.GetUserId()),
		zap.Int64("order_id", req.GetOrderId()),
		zap.String("status", statusToString(req.GetStatus())),
	)
	return &emptypb.Empty{}, nil
}
