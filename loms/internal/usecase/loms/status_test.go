package loms

import (
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	"github.com/stretchr/testify/require"
)

func TestToPortOrderStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     entity.OrderStatus
		wantStatus port.OrderStatus
	}{
		{name: "new", status: entity.OrderStatusNew, wantStatus: port.OrderStatusNew},
		{name: "awaiting payment", status: entity.OrderStatusAwaitingPayment, wantStatus: port.OrderStatusAwaitingPayment},
		{name: "failed", status: entity.OrderStatusFailed, wantStatus: port.OrderStatusFailed},
		{name: "paid", status: entity.OrderStatusPaid, wantStatus: port.OrderStatusPaid},
		{name: "cancelled", status: entity.OrderStatusCancelled, wantStatus: port.OrderStatusCancelled},
		{name: "unknown", status: entity.OrderStatus(100), wantStatus: port.OrderStatusUnspecified},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			status := toPortOrderStatus(test.status)

			require.Equal(t, test.wantStatus, status)
		})
	}
}
