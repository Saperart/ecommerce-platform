package grpc

import (
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"github.com/stretchr/testify/require"
)

func TestToProtoStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     port.OrderStatus
		wantStatus notificationspb.OrderStatus
	}{
		{
			name:       "unspecified",
			status:     port.OrderStatusUnspecified,
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
		},
		{
			name:       "new",
			status:     port.OrderStatusNew,
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_NEW,
		},
		{
			name:       "awaiting payment",
			status:     port.OrderStatusAwaitingPayment,
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT,
		},
		{
			name:       "failed",
			status:     port.OrderStatusFailed,
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_FAILED,
		},
		{
			name:       "paid",
			status:     port.OrderStatusPaid,
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_PAID,
		},
		{
			name:       "cancelled",
			status:     port.OrderStatusCancelled,
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_CANCELLED,
		},
		{
			name:       "unknown",
			status:     port.OrderStatus("unknown"),
			wantStatus: notificationspb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			status := toProtoStatus(test.status)

			require.Equal(t, test.wantStatus, status)
		})
	}
}
