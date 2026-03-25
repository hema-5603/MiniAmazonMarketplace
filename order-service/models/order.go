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
	ProductID string `json:"product_id"`
	SellerID string `json:"seller_id"`
	Quantity int `json:"quantity"`
	Price float64 `json:"price"`
}

type CheckoutRequest struct{
	Items []CheckoutItem `json:"items"`
}

