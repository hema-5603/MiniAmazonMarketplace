package repository

import (
	"database/sql"
	"product-service/models"
)

type ProductRepository interface{
	CreateProduct(product *models.Product) error
}

type productRepository struct{
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository{
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(p *models.Product) error{
	query := `INSERT INTO products (id, seller_id, name, description, price, stock, category, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?)`

	_, err := r.db.Exec(query, p.ID, p.SellerID, p.Name, p.Description, p.Price, p.Stock, p.Category, p.CreatedAt, p.UpdatedAt)
	return err
}