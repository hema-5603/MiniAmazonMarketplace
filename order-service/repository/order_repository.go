package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"order-service/models"
)

type OrderRepository interface{
	CreateOrder(ctx context.Context, order *models.Order) error
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

