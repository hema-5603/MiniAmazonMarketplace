package repository

import (
	"database/sql"
	"errors"
	"product-service/models"
)

type ProductRepository interface{
	CreateProduct(product *models.Product) error
	GetProductByID(id string) (*models.Product, error)
	UpdateProduct(*models.Product) error
	UpdateStock(productID string, newStock int) error
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

//Fetch a single product
func (r *productRepository) GetProductByID(id string) (*models.Product, error){
	query := `SELECT id, seller_id, name, description, price, stock, category, created_at, updated_at FROM products WHERE id=?`
	row := r.db.QueryRow(query,id)

	var p models.Product
	err := row.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, errors.New("Product not found")
		}
		return nil,err
	}
	return &p, nil
}

//Update an existing product
func (r *productRepository) UpdateProduct(p *models.Product)error{
	query := `UPDATE products SET name = ?, description = ?,price = ?,stock = ?,category = ?,updated_at = ? WHERE id = ?`
	_,err := r.db.Exec(query, p.Name, p.Description, p.Price, p.Stock, p.Category, p.UpdatedAt, p.ID)
	return err
}

// Update the product stock
func (r *productRepository) UpdateStock(productID string, newStock int) error{
	query := `UPDATE products SET stock = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, newStock, productID)
	return err
}