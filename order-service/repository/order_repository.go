package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"order-service/models"
)

type OrderRepository interface{
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	CountOrdersByUserID(ctx context.Context, userID string, status string) (int64, error)
	GetOrdersByUserID(ctx context.Context, userID string, status string, limit, offset int) ([]models.Order, error)
	GetPendingOrdersOlderThan(ctx context.Context, threshold time.Time) ([]string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error
}

type orderRepository struct{
	db *sql.DB
}

// NewOrderRepository acts as the constructor
func NewOrderRepository (db *sql.DB)OrderRepository{
	return &orderRepository{db: db}
}

// CreateOrder safely inserts the order and it's items using a SQL Transaction
func (r *orderRepository) CreateOrder(ctx context.Context, order *models.Order) error{
	//Extract Request ID for tracing
	reqID, _ := ctx.Value(models.RequestIDKey).(string)

	// 1. Begin the transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil{
		slog.Error("Failed to begin SQL transaction", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return errors.New("Database transaction initialization failed")
	}

	// 2. Defer a rollback
	// If the function panics or returns early with an error, this safely undoes any partial inserts.
	// If tx.Commit() succeeds at the end, this rollback does nothing

	defer func ()  {
		if p := recover(); p != nil{
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		}else if err!=nil{
			tx.Rollback()
		}
	}()

	// 3. Insert the parent order record
	orderQuery := `INSERT INTO orders (id, user_id, total_amount, status, created_at, updated_at) VALUES (?,?,?,?,?,?)`

	_, err = tx.ExecContext(ctx, orderQuery,
			order.ID,
			order.UserID,
			order.TotalAmount,
			order.Status,
			order.CreatedAt,
			order.UpdatedAt)	
	
	if err != nil{
		slog.Error("Failed to insert main order record", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return errors.New("Failed to save order record")
	}

	//4. Insert the order items (loop)
	itemQuery := `INSERT INTO order_items (id, order_id, product_id, seller_id, quantity, price, created_at) VALUES (?,?,?,?,?,?,?)`

	for _, item := range order.Items{
		_, err = tx.ExecContext(ctx, itemQuery,
				item.ID,
				order.ID,
				item.ProductID,
				item.SellerID,
				item.Quantity,
				item.Price,
				order.CreatedAt,
		)
		
		if err != nil{
			slog.Error("Failed to insert order item",
					slog.String("request_id", reqID),
					slog.String("product_id", item.ProductID),
					slog.String("error", err.Error()),
			)
			//Returning the error triggers the deferred rollback, wipes the parent order and previous items!
			return errors.New("Failed to save one or more order items")
		}
	}

	// 5. Commit the transaction
	// Since the SQL commands succeeded, write them to the disk
	err = tx.Commit()
	if err != nil{
		slog.Error("Failed to commit SQL transaction", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return errors.New("Failed to commit the order transaction")
	}

	slog.Debug("Orders and items successfully persisted to database", slog.String("request_id", reqID), slog.String("order_id", order.ID))
	return nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error){
	// 1. Fetch the parent order
	orderQuery := `SELECT id, user_id, total_amount, status, created_at, updated_at FROM orders where id = ?`

	row := r.db.QueryRowContext(ctx, orderQuery, orderID)

	var order models.Order
	err := row.Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, errors.New("Order not found")
		}
		return nil, err
	}

	// 2. Fetch the associated order items
	itemsQuery := `SELECT id, order_id, product_id, seller_id, quantity, price FROM order_items WHERE order_id = ?`
	
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next(){
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.SellerID, &item.Quantity, &item.Price)
		if err != nil{
			return nil, err
		}
		items = append(items, item)
	}

	// 3. Attach the item to the parent order
	order.Items = items

	return &order, nil
}

func (r *orderRepository) CountOrdersByUserID(ctx context.Context, userID string, status string) (int64, error){
	query := `SELECT COUNT(*) FROM orders WHERE user_id = ? `
	args := []interface{}{userID}

	if status != ""{
		query += `AND status = ? `
		args = append(args, status)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID string, status string, limit, offset int) ([]models.Order, error){
	query := `SELECT id, user_id, total_amount, status, created_at, updated_at FROM orders WHERE user_id = ? `
	args := []interface{}{userID}

	//Dynamically append the status filter if the user provided one
	if status != ""{
		query += `AND status = ? `
		args = append(args, status)
	}

	// Always order by newest first
	query += `ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil{
		return nil, err
	}

	defer rows.Close()

	var orders []models.Order
	for rows.Next(){
		var o models.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil{
			return nil, err
		}

		// Initialize empty items array so it returns `[]` in JSON instead of `null`
		o.Items = []models.OrderItem{}
		orders = append(orders, o)
	}
	return orders, nil
}


func (r *orderRepository) GetPendingOrdersOlderThan(ctx context.Context, threshold time.Time) ([]string, error){
	
	slog.Info("Cron: Querying for orders older than", slog.Time("threshold",threshold))
	
	query := `SELECT id, created_at FROM orders WHERE status = ?`
	rows, err := r.db.QueryContext(ctx, query, models.StatusPending)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	var orderIDs []string
	for rows.Next(){
		var id string
		var createdAtBytes []byte 

		// 1. Scan the raw data, If it fails, log the real error and skip to the next row
		if err := rows.Scan(&id, &createdAtBytes); err != nil{
			slog.Error("Failed to scan row", slog.String("error", err.Error()))
			continue
		}
		// 2. Convert the raw bytes to a string
		dateStr := string(createdAtBytes)
		var parsedTime time.Time

		// 3. Try parsing it using the ISO format
		parsedTime, err = time.Parse(time.RFC3339, dateStr)
		if err != nil{
			parsedTime, _ = time.Parse("2006-01-02 15:04:05", dateStr)
		}
		// 4. Mathematical comparison
		if parsedTime.Before(threshold){
			slog.Info("Successfully caught an expired order",
				slog.String("order_id", id),
				slog.Time("Order_time", parsedTime),
			)
			orderIDs = append(orderIDs, id)
		}
	}
	return orderIDs, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error{
	query := `UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil{
		return err
	}

	// Safety check to ensure we actually updated a row
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0{
		return errors.New("Order not found or status already updated")
	}
	return nil
}