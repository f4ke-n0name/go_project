package models

import "time"

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusConfirmed  OrderStatus = "CONFIRMED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ID       string  `json:"id"`
	OrderID  string  `json:"order_id"`
	PartID   string  `json:"part_id"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

type Order struct {
	ID         string       `json:"id"`
	UserID     string       `json:"user_id"`
	Status     OrderStatus  `json:"status"`
	TotalPrice float64      `json:"total_price"`
	Items      []*OrderItem `json:"items"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type OutboxEvent struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
