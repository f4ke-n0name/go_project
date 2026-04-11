package models

import "time"

type MovementType string

const (
	MovementTypeReserve  MovementType = "reserve"
	MovementTypeRelease  MovementType = "release"
	MovementTypeRestock  MovementType = "restock"
	MovementTypeWriteoff MovementType = "writeoff"
)

type Stock struct {
	ID        string
	PartID    string
	Quantity  int32
	Reserved  int32
	UpdatedAt time.Time
}

func (s *Stock) Available() int32 {
	return s.Quantity - s.Reserved
}

type StockMovement struct {
	ID        string
	PartID    string
	OrderID   string
	Type      MovementType
	Quantity  int32
	CreatedAt time.Time
}

type ReserveItem struct {
	PartID   string
	Quantity int32
}
