package postgres

import (
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	sqlcorder "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/postgres/sqlc"
	"github.com/stretchr/testify/require"
)

func TestToSQLCStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     entity.OrderStatus
		wantStatus sqlcorder.LomsOrderStatus
	}{
		{name: "new", status: entity.OrderStatusNew, wantStatus: sqlcorder.LomsOrderStatusNew},
		{name: "awaiting payment", status: entity.OrderStatusAwaitingPayment, wantStatus: sqlcorder.LomsOrderStatusAwaitingPayment},
		{name: "failed", status: entity.OrderStatusFailed, wantStatus: sqlcorder.LomsOrderStatusFailed},
		{name: "paid", status: entity.OrderStatusPaid, wantStatus: sqlcorder.LomsOrderStatusPaid},
		{name: "cancelled", status: entity.OrderStatusCancelled, wantStatus: sqlcorder.LomsOrderStatusCancelled},
		{name: "unknown", status: entity.OrderStatus(100), wantStatus: sqlcorder.LomsOrderStatusNew},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			status := toSQLCStatus(test.status)

			require.Equal(t, test.wantStatus, status)
		})
	}
}

func TestFromSQLCStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     sqlcorder.LomsOrderStatus
		wantStatus entity.OrderStatus
	}{
		{name: "new", status: sqlcorder.LomsOrderStatusNew, wantStatus: entity.OrderStatusNew},
		{name: "awaiting payment", status: sqlcorder.LomsOrderStatusAwaitingPayment, wantStatus: entity.OrderStatusAwaitingPayment},
		{name: "failed", status: sqlcorder.LomsOrderStatusFailed, wantStatus: entity.OrderStatusFailed},
		{name: "paid", status: sqlcorder.LomsOrderStatusPaid, wantStatus: entity.OrderStatusPaid},
		{name: "cancelled", status: sqlcorder.LomsOrderStatusCancelled, wantStatus: entity.OrderStatusCancelled},
		{name: "unknown", status: sqlcorder.LomsOrderStatus("unknown"), wantStatus: entity.OrderStatusUnspecified},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			status := fromSQLCStatus(test.status)

			require.Equal(t, test.wantStatus, status)
		})
	}
}
