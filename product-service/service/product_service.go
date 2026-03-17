package service

import (
	"errors"
	"log/slog"
	"time"

	"product-service/models"
	"product-service/repository"

	"github.com/google/uuid"
)

type ProductService interface{
	CreateProduct(sellerID string, req models.CreateProductRequest) (*models.Product, error)
	UpdateProduct(productID string, sellerID string, req models.UpdateProductRequest) (*models.Product, error)
}

type productService struct{
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService{
	return &productService{repo:repo}
}

func (s *productService) CreateProduct(sellerID string, req models.CreateProductRequest) (*models.Product, error){
	if req.Name == "" || req.Price <= 0{
		slog.Warn("Product creation failed: Invalid data", slog.String("seller_id", sellerID))
		return nil, errors.New("Invalid product data: Name and Positive price is required")
	}

	product := &models.Product{
		ID: uuid.New().String(),
		SellerID: sellerID,
		Name: req.Name,
		Description: req.Description,
		Price: req.Price,
		Stock: req.Stock,
		Category: req.Category,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.CreateProduct(product)

	if err != nil{
		slog.Error("Database error during product creation", slog.String("error",err.Error()))
		return nil, errors.New("Failed to create product")
	}

	slog.Info("Product created successfully", slog.String("product_id", product.ID), slog.String("seller_id", sellerID))

	return product, nil
}

func (s* productService) UpdateProduct(productID string, sellerID string, req models.UpdateProductRequest) (*models.Product, error){
	// 1. Fetch the existing product
	product, err := s.repo.GetProductByID(productID)
	if err != nil{
		return nil, err
	}

	// 2. Resource ownership check
	if product.SellerID != sellerID{
		slog.Warn("Unauthorized product update attempt",
				slog.String("product_id",productID),
				slog.String("attempted_by",sellerID),
				slog.String("actual_owner",product.SellerID),
		)
		return nil, errors.New("Unauthorized: You do not own this product")
	}

	// 3. Validate new data
	if req.Name == "" || req.Price <= 0{
		return nil,errors.New("Invalid product data: Name and positive price are required")
	} 

	// 4. Apply updates
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.Category = req.Category
	product.UpdatedAt = time.Now()

	// 5. Save to database
	err = s.repo.UpdateProduct(product)
	if err != nil{
		slog.Error("Database error during product update", slog.String("error", err.Error()))
		return nil,errors.New("Failed to update product")
	}
	slog.Info("Product updated successfully", slog.String("product_id",product.ID))
	return product, nil
}