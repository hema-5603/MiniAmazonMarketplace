package models

import "time"

type ContextKey string
const RequestIDKey ContextKey = "request_id"

type OrderStatus string
const(
	StatusPending OrderStatus = "PENDING" // Waiting for payment/stock reservation
	StatusPaid OrderStatus = "PAID" // Checkout complete
	StatusCancelled OrderStatus = "CANCELLED" // Cancelled by user
	StatusExpired OrderStatus = "EXPIRED" // Cron job killed it because they took too long
)

type OrderItem struct{
	ID string `json:"id"`
	OrderID string `json:"order_id"`
	ProductID string `json:"product_id"`
	SellerID string `json:"seller_id"`
	Quantity int `json:"quantity"`
	Price float64 `json:"price"`
}

type Order struct{
	ID string `json:"id"`
	UserID string `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status OrderStatus `json:"status"`
	Items []OrderItem `json:"items"` //Nested 
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CheckoutItem struct{
	ProductID string `json:"product_id" validate:"required"`
	SellerID string `json:"seller_id"`
	Quantity int `json:"quantity" validate:"required,min=1"`
	Price float64 `json:"price" validate:"required,gt=0"`
}

type CheckoutRequest struct{
	Items []CheckoutItem `json:"items" validate:"required,min=1,dive"` // Dive ensures validation of the items inside the array too
}

// Pagination meta information
type PaginatedMeta struct{
	CurrentPage int `json:"current_page"`
	PageSize int `json:"page_size"`
	TotalItems int 	`json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// Paginated response wrapper for order history
type PaginatedOrderResponse struct{
	Data []Order `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}