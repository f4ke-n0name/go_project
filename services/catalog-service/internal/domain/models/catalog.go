package models

import "time"

type AutoPart struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Brand       string    `json:"brand"`
	Price       float64   `json:"price"`
	CategoryID  string    `json:"category_id"`
	Stock       int32     `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  string    `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
}

type Compatibility struct {
	ID       string `json:"id"`
	PartID   string `json:"partId"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	YearFrom int32  `json:"yearFrom"`
	YearTo   int32  `json:"yearTo"`
}
