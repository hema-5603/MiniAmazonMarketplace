package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"order-service/client"
	"order-service/models"
	"order-service/repository"

	"github.com/google/uuid"
)

type OrderService interface{
	CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error)
	GetOrderDetail(ctx context.Context, orderID string ,userID string) (*models.Order, error)
	GetOrderHistory(ctx context.Context, userID string, page, limit int, status string) (*models.PaginatedOrderResponse, error)
	ExpireUnpaidOrders(ctx context.Context) error
	CancelOrder(ctx context.Context, orderID string, userID string) error
}

type orderService struct{
	repo repository.OrderRepository
	productClient client.ProductClient
}

func NewOrderService(repo repository.OrderRepository, productClient client.ProductClient) OrderService{
	return &orderService{
		repo: repo,
		productClient: productClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error){
	reqID, _ := ctx.Value(models.RequestIDKey).(string) 
	
	if len(req.Items) == 0 {
		slog.Warn("Checkout attempted with empty cart", slog.String("request_id", reqID), slog.String("user_id", userID))
		return nil, errors.New("Couldn't create an order with an empty cart")
	}

	// 1. Calling the product service to validate the cart
	validationResp, err := s.productClient.ValidateCart(ctx, req.Items)
	if err != nil{
		return nil, err
	}

	// 2. Hard block if any items is out of stock or deactivated
	if !validationResp.AllAvailable{
		slog.Warn("Checkout rejected: Item unavailable", slog.String("request_id", reqID), slog.String("user_id", userID))
		return nil, errors.New("One or more items in your cart are out of stock or unavailable")
	}

	// 3. Create a map of the TRUE prices and seller IDs from the product service
	truthMap :=  make(map[string]struct{
		Price float64
		SellerID string
	})
	for _, item := range validationResp.Data{
		truthMap[item.ProductID] = struct{
			Price float64 
			SellerID string
		}{item.Price, item.SellerID}
	}

	// 2. Initialize the order
	orderID := uuid.New().String()
	var totalAmount float64
	var orderItems []models.OrderItem
	now := time.Now()

	// 2. Process each item and calculate the total amount
	for _, reqItem := range req.Items{
		trueData, exists := truthMap[reqItem.ProductID]
		if !exists{
			slog.Warn("Product validation mismatch", slog.String("request_id", reqID), slog.String("product_id", reqItem.ProductID))
			return nil, errors.New("Mismatch in product validation")
		}

		// The price mismatch check
		if reqItem.Price != trueData.Price{
			slog.Warn("Price mismatched during checkout",
				slog.String("request_id", reqID),
				slog.String("product_id", reqItem.ProductID),
				slog.Float64("cart_price", reqItem.Price),
				slog.Float64("actual_price", trueData.Price),
			)
			return nil, errors.New("The Price of one or more items has changed.")
		}
		lineItemTotal := trueData.Price * float64(reqItem.Quantity)
		totalAmount += lineItemTotal

		orderItems = append(orderItems, models.OrderItem{
			ID: uuid.New().String(),
			OrderID: orderID,
			ProductID: reqItem.ProductID,
			SellerID: trueData.SellerID,
			Quantity: reqItem.Quantity,
			Price: trueData.Price,
		})
	}

	// 3. Build the final order object
	order := &models.Order{
		ID: orderID,
		UserID: userID,
		TotalAmount: totalAmount,
		Status: models.StatusPending, // Always starts as pending
		Items: orderItems,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Reserve the inventory
	err = s.productClient.ReserveStock(ctx, req.Items)
	if err != nil{
		// The product service blocked it
		slog.Warn("Checkout blocked during reservation", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return nil, errors.New("Checkout failed during inventory reservation: " + err.Error())
	}
	// 4. Call the repository which triggers the SQL transaction
	err = s.repo.CreateOrder(ctx, order)
	if err != nil{
		slog.Error("Failed to save order after reserving stock", slog.String("request_id", reqID), slog.String("order_id",orderID), slog.String("error", err.Error()))
		return nil , errors.New("System error occurred while finalizing your order")
	}

	slog.Info("Order created successfully", 
			 slog.String("request_id", reqID), 
			 slog.String("order_id",orderID),
			 slog.String("user_id", userID),
			)
	return order, nil
}

func (s *orderService) GetOrderDetail(ctx context.Context, orderID string, userID string) (*models.Order, error){
	reqID, _ := ctx.Value(models.RequestIDKey).(string)

	// Fetch the product from the database 
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil{
		if err.Error() == "Order not found" {
			return nil, err // Pass 404
		}
		slog.Error("Database error fetching order", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return nil, errors.New("Failed to retrieve order")
	}

	// Resource ownership check
	// Block the user who tries to guess another user's orderID
	if order.UserID != userID{
		slog.Warn("Unauthorized order access attempt",
				slog.String("request_id", reqID),
				slog.String("attempted_by_user", userID),
				slog.String("order_owner", order.UserID),
			)
			return nil, errors.New("Unauthorized: You do not own this order")
	}

	slog.Debug("Order detail fetched successfully", slog.String("request_id", reqID), slog.String("order_id", orderID))
	return order, nil
}

func (s *orderService) GetOrderHistory(ctx context.Context, userID string, page, limit int, status string) (*models.PaginatedOrderResponse, error){
	reqID, _ := ctx.Value(models.RequestIDKey).(string)

	// 1. Sanitize pagination inputs
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50{
		limit = 10
	}

	offset := (page - 1) * limit

	// 2. Execute DB queries
	totalItems, err := s.repo.CountOrdersByUserID(ctx, userID, status)
	if err != nil{
		slog.Error("Failed to count orders", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return nil, errors.New("Failed to retrieve order history")
	}

	orders, err := s.repo.GetOrdersByUserID(ctx, userID, status, limit, offset)
	if err != nil{
		slog.Error("Failed to fetch order history", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return nil, errors.New("Failed to retrieve order history")
	}

	// Prevent returning `null`in JSON if the user has no orders
	if orders == nil{
		orders = []models.Order{}
	}

	// 3. Calculate total pages (Ceiling division trick in Go)
	totalPages := int((totalItems + int64(limit) - 1) /int64(limit))

	res := &models.PaginatedOrderResponse{
		Data: orders,
		Meta: models.PaginatedMeta{
			CurrentPage: page,
			PageSize: limit,
			TotalItems: int(totalItems),
			TotalPages: totalPages,
		},
	}

	slog.Debug("Order history fetched successfully", slog.String("request_id", reqID), slog.String("user_id", userID))
	return res, nil
}

func (s *orderService) ExpireUnpaidOrders(ctx context.Context) error{
	//PENDING order older than 15 minutes would be expire
	expirationTime := time.Now().Add(-15*time.Minute)

	slog.Info("Cron: Searching for orders older than", slog.Time("threshold", expirationTime))
	// 1. Find the expired orders
	orderIDs, err := s.repo.GetPendingOrdersOlderThan(ctx, expirationTime)
	if err != nil{
		slog.Error("Cron: Failed to fetch expired orders", slog.String("error", err.Error()))
		return err
	}

	if len(orderIDs) == 0{
		return nil // Nothing to do
	}

	slog.Info("Cron: Found unpaid orders to expire", slog.Int("count", len(orderIDs)))

	// 2. Process each expired orders
	for _, id := range orderIDs{
		// Fetch the full order details to know what time to release
		order, err := s.repo.GetOrderByID(ctx,id)
		if err != nil{
			continue
		}

		// Tell the orders to put the products back on the stock
		err = s.productClient.ReleaseStock(ctx, order.Items)
		if err != nil{
			slog.Error("Cron: Failed to release stock", slog.String("order_id", id))
			continue // Skip updating the status
		}
		
		// Mark the order as EXPIRED in the database
		err = s.repo.UpdateOrderStatus(ctx, id, models.StatusExpired)
		if err != nil{
			slog.Error("Cron: Failed to update order status", slog.String("order_id", id))
		}else {
			slog.Info("Cron: Successfully expired order and released stock", slog.String("order_id", id))
		}
	}
	return nil
}

func (s *orderService) CancelOrder(ctx context.Context, orderID string, userID string) error{
	reqID, _ := ctx.Value(models.RequestIDKey).(string)

	// 1. Fetch the full order(which includes the items needed for restock)
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil{
		return errors.New("Order not found")
	}

	// 2. Prevent users from cancelling other people's orders
	if order.UserID != userID{
		slog.Warn("Unauthorized cancellation attempt", slog.String("request_id", reqID), slog.String("user_id", userID))
		return errors.New("Unauthorized: You do not own this product")
	}

	// 3. Business logic: You can cancel a PENDING order
	if order.Status != models.StatusPending{
		slog.Warn("Attempted to cancel non-pending order", slog.String("request_id", reqID), slog.String("status",string(order.Status)))
		return errors.New("Only pending orders can be cancelled")
	}

	// 4. RESTOCK: Tell the product service to put the items back on the shelf
	err = s.productClient.ReleaseStock(ctx, order.Items)
	if err != nil{
		slog.Error("Failed to release stock during cancellation", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return errors.New("Failed to communicate with inventory system")
	}

	// 5. Update the database to CANCELLED
	err = s.repo.UpdateOrderStatus(ctx, orderID, models.StatusCancelled)
	if err != nil{
		slog.Error("Failed to update order status to cancelled", slog.String("request_id", reqID))
		return errors.New("System error occurred while cancelling the order")
	}

	slog.Info("Order successfully cancelled manually", slog.String("request_id", reqID), slog.String("order_id", orderID))
	return nil
}