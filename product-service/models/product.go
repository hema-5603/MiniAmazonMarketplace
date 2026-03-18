package models

import "time"

type Product struct{
	ID string `json:"id"`
	SellerID string `json:"seller_id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Price float64 `json:"price"`
	Stock int `json:"stock"`
	Category string `json:"category"`
	IsActive bool `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct{
	Name string `json:"name"`
	Description string `json:"description"`
	Price float64 `json:"price"`
	Stock int `json:"stock"`
	Category string `json:"category"`
}

type UpdateProductRequest struct{
	Name string `json:"name"`
	Description string `json:"description"`
	Price float64 `json:"price"`
	Stock int `json:"stock"`
	Category string `json:"category"`
}

type UpdateProductStockRequest struct{
	Stock int `json:"stock"`
}