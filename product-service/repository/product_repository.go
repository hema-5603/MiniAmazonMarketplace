package repository

import (
	"context"
	"database/sql"
	"errors"
	"product-service/models"
)

type ProductRepository interface{
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	UpdateProduct(ctx context.Context,p *models.Product) error
	UpdateStock(ctx context.Context, productID string, newStock int) error
	UpdateProductStatus(ctx context.Context, productID string, isActive bool) error

	GetProducts(ctx context.Context, limit, offset int, search, category string) ([]models.Product,error)
	CountProducts(ctx context.Context, search, category string) (int64, error)

	ReserveStock(ctx context.Context, items []models.ReserveItem) error
	ReleaseStock(ctx context.Context, items []models.ReserveItem) error
}

type productRepository struct{
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository{
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(ctx context.Context, p *models.Product) error{
	query := `INSERT INTO products (id, seller_id, name, description, price, stock, category, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?)`

	_, err := r.db.ExecContext(ctx,query, p.ID, p.SellerID, p.Name, p.Description, p.Price, p.Stock, p.Category, p.CreatedAt, p.UpdatedAt)
	return err
}

//Fetch a single product
func (r *productRepository) GetProductByID(ctx context.Context, id string) (*models.Product, error){
	query := `SELECT id, seller_id, name, description, price, stock, category, is_active, created_at, updated_at FROM products WHERE id=?`
	row := r.db.QueryRowContext(ctx,query,id)

	var p models.Product
	err := row.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category,&p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, errors.New("Product not found")
		}
		return nil,err
	}
	return &p, nil
}

//Update an existing product
func (r *productRepository) UpdateProduct(ctx context.Context, p *models.Product)error{
	query := `UPDATE products SET name = ?, description = ?,price = ?,stock = ?,category = ?,updated_at = ? WHERE id = ?`
	_,err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.Price, p.Stock, p.Category, p.UpdatedAt, p.ID)
	return err
}

// Update the product stock
func (r *productRepository) UpdateStock(ctx context.Context, productID string, newStock int) error{
	query := `UPDATE products SET stock = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, newStock, productID)
	return err
}

// Deactivate the product
func (r *productRepository) UpdateProductStatus(ctx context.Context, productID string, isActive bool) error{
	query := `UPDATE products SET is_active = ?, updated_at = NOW() where id = ?`
	_, err := r.db.ExecContext(ctx, query,isActive, productID)
	return err
}

// Dynamic data fetcher
func (r *productRepository) GetProducts(ctx context.Context, limit, offset int, search, category string) ([]models.Product,error){
	// Enforce is_active = TRUE for the public catalog
	query := `SELECT id, seller_id, name, description, price, stock, category, is_active, created_at, updated_at FROM products WHERE is_active = TRUE`

	var args []interface{}

	if search != ""{
		query += `AND name LIKE ?`
		args = append(args, "%"+search+"%") //Enables partial matching the user prompted name with the product name
	}
	if category != ""{
		query += `AND category = ?`
		args = append(args, category)
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next(){
		var p models.Product
		err := rows.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil{
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// Counter for total pages
func (r *productRepository) CountProducts(ctx context.Context, search, category string) (int64, error){
	query := `SELECT COUNT(*) FROM products WHERE is_active = TRUE`
	var args []interface{}

	if search != ""{
		query += `AND name LIKE ?`
		args = append(args, "%"+search+"%")
	}
	if category != ""{
		query += `AND category = ?`
		args = append(args, category)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *productRepository) ReserveStock(ctx context.Context, items []models.ReserveItem) error{
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil{
		return err
	}
	defer tx.Rollback() //safely net

	query := `UPDATE products SET stock = stock - ?, updated_at = NOW() WHERE id = ? AND stock >= ? AND is_active = TRUE`

	for _, item := range items{
		res, err := tx.ExecContext(ctx, query, item.Quantity, item.ProductID, item.Quantity)
		if err != nil{
			return err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil{
			return err
		}

		// If 0 rows are updated, that means the product doesn't exist or inactive or stock < requested quantity
		if rowsAffected == 0{
			return errors.New("Failed to reserve stock: Insufficient inventory for product " + item.ProductID)
		}
	}

	return tx.Commit()
}

func (r *productRepository) ReleaseStock(ctx context.Context, items []models.ReserveItem) error{
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil{
		return err
	}
	defer tx.Rollback()

	// Adding back the stock
	query := `UPDATE products SET stock = stock + ?, updated_at = NOW() where id = ?`

	for _, item := range items{
		_, err := tx.ExecContext(ctx, query, item.Quantity, item.ProductID)
		if err != nil{
			return err
		}
	}

	return tx.Commit()
}