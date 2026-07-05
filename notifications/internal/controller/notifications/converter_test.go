package notifications

import (
	"testing"

	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"github.com/stretchr/testify/require"
)

func TestStatusToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status notificationspb.OrderStatus
		want   string
	}{
		{
			name:   "new",
			status: notificationspb.OrderStatus_ORDER_STATUS_NEW,
			want:   "new",
		},
		{
			name:   "awaiting payment",
			status: notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT,
			want:   "awaiting_payment",
		},
		{
			name:   "failed",
			status: notificationspb.OrderStatus_ORDER_STATUS_FAILED,
			want:   "failed",
		},
		{
			name:   "paid",
			status: notificationspb.OrderStatus_ORDER_STATUS_PAID,
			want:   "paid",
		},
		{
			name:   "cancelled",
			status: notificationspb.OrderStatus_ORDER_STATUS_CANCELLED,
			want:   "cancelled",
		},
		{
			name:   "unspecified",
			status: notificationspb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
			want:   "unspecified",
		},
		{
			name:   "unknown enum value",
			status: notificationspb.OrderStatus(999),
			want:   "unspecified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := statusToString(test.status)

			require.Equal(t, test.want, got)
		})
	}
}
