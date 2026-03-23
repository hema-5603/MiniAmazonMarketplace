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

// Product status
type UpdateStatusRequest struct{
	IsActive bool `json:"is_active"`
}

// Get product list
type PaginationMeta struct{
	CurrentPage int `json:"current_page"`
	TotalPages int `json:"total_pages"`
	PageSize int `json:"page_size"`
	TotalItems int64 `json:"total_items"`
}

type PaginatedProductResponse struct{
	Data []Product `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
// Item structure received in the request
type StockCheckItem struct{
	ProductID string `json:"product_id"`
	RequestedQuantity int `json:"requested_quantity"`
}

//Overall request payload
type StockCheckRequest struct{
	Items []StockCheckItem `json:"items"`
}

//Detailed result for each item
type StockCheckResult struct{
	ProductID string `json:"product_id"`
	HasStock bool `json:"has_stock"`
	CurrentStock int `json:"current_stock"`
	Message string `json:"message"`
}