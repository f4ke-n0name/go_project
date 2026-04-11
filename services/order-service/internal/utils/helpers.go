package utils

import (
	"github.com/f4ke-n0name/autoparts-hub/gen/order"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/domain/models"
)

func ToProtoOrder(o *models.Order) *order.Order {
	items := make([]*order.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, &order.OrderItem{
			PartId:   item.PartID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	return &order.Order{
		Id:         o.ID,
		UserId:     o.UserID,
		Status:     modelStatusToProto(o.Status),
		TotalPrice: o.TotalPrice,
		Items:      items,
		CreatedAt:  o.CreatedAt.String(),
	}
}

func modelStatusToProto(s models.OrderStatus) order.OrderStatus {
	switch s {
	case models.OrderStatusPending:
		return order.OrderStatus_ORDER_STATUS_PENDING
	case models.OrderStatusConfirmed:
		return order.OrderStatus_ORDER_STATUS_CONFIRMED
	case models.OrderStatusProcessing:
		return order.OrderStatus_ORDER_STATUS_PROCESSING
	case models.OrderStatusShipped:
		return order.OrderStatus_ORDER_STATUS_SHIPPED
	case models.OrderStatusDelivered:
		return order.OrderStatus_ORDER_STATUS_DELIVERED
	case models.OrderStatusCancelled:
		return order.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return order.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func ProtoStatusToModel(s order.OrderStatus) models.OrderStatus {
	switch s {
	case order.OrderStatus_ORDER_STATUS_PENDING:
		return models.OrderStatusPending
	case order.OrderStatus_ORDER_STATUS_CONFIRMED:
		return models.OrderStatusConfirmed
	case order.OrderStatus_ORDER_STATUS_PROCESSING:
		return models.OrderStatusProcessing
	case order.OrderStatus_ORDER_STATUS_SHIPPED:
		return models.OrderStatusShipped
	case order.OrderStatus_ORDER_STATUS_DELIVERED:
		return models.OrderStatusDelivered
	case order.OrderStatus_ORDER_STATUS_CANCELLED:
		return models.OrderStatusCancelled
	default:
		return models.OrderStatusPending
	}
}
