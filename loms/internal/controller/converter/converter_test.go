package converter

import (
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
)

func TestProtoToItemsOrder(t *testing.T) {
	t.Parallel()

	items := ProtoToItemsOrder([]*lomspb.Item{
		{Sku: 10, Count: 2},
		{Sku: 12, Count: 1},
	})

	require.Equal(t, []entity.OrderItem{
		{SKU: 10, Count: 2},
		{SKU: 12, Count: 1},
	}, items)
}

func TestItemsOrderToProto(t *testing.T) {
	t.Parallel()

	items := ItemsOrderToProto([]entity.OrderItem{
		{SKU: 10, Count: 2},
		{SKU: 12, Count: 1},
	})

	require.Equal(t, []*lomspb.Item{
		{Sku: 10, Count: 2},
		{Sku: 12, Count: 1},
	}, items)
}

func TestOrderToProto(t *testing.T) {
	t.Parallel()

	createdAt := time.Unix(10, 0)
	updatedAt := time.Unix(20, 0)

	tests := []struct {
		name      string
		order     *entity.Order
		wantOrder *lomspb.GetOrderResponse
	}{
		{
			name:      "nil order",
			order:     nil,
			wantOrder: nil,
		},
		{
			name: "success",
			order: &entity.Order{
				ID:        1001,
				UserID:    42,
				Status:    entity.OrderStatusPaid,
				Items:     []entity.OrderItem{{SKU: 10, Count: 2}},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			wantOrder: &lomspb.GetOrderResponse{
				Status:    lomspb.OrderStatus_ORDER_STATUS_PAID,
				UserId:    42,
				Items:     []*lomspb.Item{{Sku: 10, Count: 2}},
				CreatedAt: nil,
				UpdatedAt: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			order := OrderToProto(test.order)

			if test.wantOrder == nil {
				require.Nil(t, order)
				return
			}

			require.Equal(t, test.wantOrder.Status, order.Status)
			require.Equal(t, test.wantOrder.UserId, order.UserId)
			require.Equal(t, test.wantOrder.Items, order.Items)
			require.Equal(t, createdAt.Unix(), order.CreatedAt.AsTime().Unix())
			require.Equal(t, updatedAt.Unix(), order.UpdatedAt.AsTime().Unix())
		})
	}
}

func TestOrderStatusToProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     entity.OrderStatus
		wantStatus lomspb.OrderStatus
	}{
		{
			name:       "unspecified",
			status:     entity.OrderStatusUnspecified,
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
		},
		{
			name:       "new",
			status:     entity.OrderStatusNew,
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_NEW,
		},
		{
			name:       "awaiting payment",
			status:     entity.OrderStatusAwaitingPayment,
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT,
		},
		{
			name:       "failed",
			status:     entity.OrderStatusFailed,
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_FAILED,
		},
		{
			name:       "paid",
			status:     entity.OrderStatusPaid,
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_PAID,
		},
		{
			name:       "cancelled",
			status:     entity.OrderStatusCancelled,
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_CANCELLED,
		},
		{
			name:       "unknown",
			status:     entity.OrderStatus(100),
			wantStatus: lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			status := OrderStatusToProto(test.status)

			require.Equal(t, test.wantStatus, status)
		})
	}
}
