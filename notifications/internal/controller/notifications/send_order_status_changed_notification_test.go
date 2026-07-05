package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNotificationsServer_SendOrderStatusChangedNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *notificationspb.OrderStatusChangedNotificationRequest
		handler     http.HandlerFunc
		emptyURL    bool
		wantPayload *callbackPayload
		wantCode    codes.Code
		wantCalls   int
	}{
		{
			name: "success",
			req: &notificationspb.OrderStatusChangedNotificationRequest{
				UserId:  1001,
				OrderId: 2002,
				Status:  notificationspb.OrderStatus_ORDER_STATUS_PAID,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusNoContent)
			},
			wantPayload: &callbackPayload{
				UserID:  1001,
				OrderID: 2002,
				Status:  "paid",
			},
			wantCode:  codes.OK,
			wantCalls: 1,
		},
		{
			name: "empty callback addr",
			req: &notificationspb.OrderStatusChangedNotificationRequest{
				UserId:  1001,
				OrderId: 2002,
				Status:  notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT,
			},
			emptyURL:  true,
			wantCode:  codes.OK,
			wantCalls: 0,
		},
		{
			name: "invalid request",
			req: &notificationspb.OrderStatusChangedNotificationRequest{
				UserId:  0,
				OrderId: 2002,
				Status:  notificationspb.OrderStatus_ORDER_STATUS_PAID,
			},
			wantCode:  codes.InvalidArgument,
			wantCalls: 0,
		},
		{
			name: "callback error",
			req: &notificationspb.OrderStatusChangedNotificationRequest{
				UserId:  1001,
				OrderId: 2002,
				Status:  notificationspb.OrderStatus_ORDER_STATUS_CANCELLED,
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			wantCode:  codes.Unavailable,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if tt.wantPayload != nil {
					var payload callbackPayload
					require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
					require.Equal(t, *tt.wantPayload, payload)
				}
				if tt.handler != nil {
					tt.handler(w, r)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			cfg := &config.Config{}
			cfg.Callback.Timeout = time.Second
			if !tt.emptyURL {
				cfg.Callback.Addr = server.URL
			}

			s := NewNotificationsServer(zap.NewNop(), cfg)
			_, err := s.SendOrderStatusChangedNotification(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			require.Equal(t, tt.wantCalls, calls)
		})
	}
}
