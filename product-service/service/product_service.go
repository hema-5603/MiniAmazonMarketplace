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