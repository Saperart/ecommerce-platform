package converter

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProtoToItemsOrder(items []*lomspb.Item) []entity.OrderItem {
	result := make([]entity.OrderItem, 0, len(items))
	for _, item := range items {
		currentOrderItem := entity.OrderItem{
			SKU:   item.GetSku(),
			Count: item.GetCount(),
		}
		result = append(result, currentOrderItem)
	}
	return result
}

func ItemsOrderToProto(items []entity.OrderItem) []*lomspb.Item {
	result := make([]*lomspb.Item, 0, len(items))
	for _, item := range items {
		currentItem := &lomspb.Item{
			Sku:   item.SKU,
			Count: item.Count,
		}
		result = append(result, currentItem)
	}
	return result
}

func OrderToProto(order *entity.Order) *lomspb.GetOrderResponse {
	if order == nil {
		return nil
	}
	return &lomspb.GetOrderResponse{
		Status:    OrderStatusToProto(order.Status),
		UserId:    order.UserID,
		Items:     ItemsOrderToProto(order.Items),
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}
}

func OrderStatusToProto(status entity.OrderStatus) lomspb.OrderStatus {
	switch status {
	case entity.OrderStatusNew:
		return lomspb.OrderStatus_ORDER_STATUS_NEW
	case entity.OrderStatusAwaitingPayment:
		return lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case entity.OrderStatusFailed:
		return lomspb.OrderStatus_ORDER_STATUS_FAILED
	case entity.OrderStatusPaid:
		return lomspb.OrderStatus_ORDER_STATUS_PAID
	case entity.OrderStatusCancelled:
		return lomspb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
