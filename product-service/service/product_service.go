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
	UpdateStock(productID string, sellerID string, req models.UpdateProductStockRequest) error
	UpdateProductStatus(productID string, sellerID string, req models.UpdateStatusRequest) error

	GetProducts(page, limit int, search, category string) (*models.PaginatedProductResponse, error)
	GetProductDetail(productID string) (*models.Product, error)
	ValidateStock(req models.StockCheckRequest) ([]models.StockCheckResult, bool, error)
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

// Service for updating stock
func (s *productService) UpdateStock(productID string, sellerID string, req models.UpdateProductStockRequest) error{
	// 1. Prevent negative stocking
	if req.Stock < 0 {
		slog.Warn("Invalid stock update attempt", slog.Int("attempted_stock",req.Stock))
		return errors.New("Invalid operation: Stock cannot be negative")
	}

	// 2. Fetch the product to check ownership
	product, err := s.repo.GetProductByID(productID)
	if err != nil{
		return err
	}

	// 3. Resouce ownership check
	if product.SellerID != sellerID{
		slog.Warn("Unauthorized stock update attempt",
				slog.String("product_id",productID),
				slog.String("attempted_by", sellerID),
				slog.String("actual_owner",product.SellerID),
		)
		return errors.New("Unauthorized: You do not own this product")
	}

	// 4. Update the stock
	err = s.repo.UpdateStock(productID, req.Stock)
	if err != nil{
		slog.Error("Database error during stock update",slog.String("error",err.Error()))
		return errors.New("Failed to update stock")
	}

	slog.Info("Stock updated successfully", slog.String("product_id",productID), slog.Int("new_stock",req.Stock))
	return nil
}

func (s *productService) UpdateProductStatus(productID string, sellerID string, req models.UpdateStatusRequest) error{
	// 1. Fetch to check ownership
	product, err := s.repo.GetProductByID(productID)
	if err != nil{
		return err
	}

	// 2. Resource ownership check
	if product.SellerID != sellerID{
		slog.Warn("Unauthorized status update attempt",
			slog.String("product_id", productID),
			slog.String("attempted_by",sellerID),
		)
		return errors.New("Unauthorized: You do not own this product")
	}

	// 3. Update status
	err = s.repo.UpdateProductStatus(productID, req.IsActive)
	if err != nil{
		slog.Error("Database error during status update", slog.String("error", err.Error()))
		return errors.New("Failed to update product status")
	}
	slog.Info("Product status updated", slog.String("product_id",productID), slog.Bool("is_active", req.IsActive))
	return nil
}

func (s *productService) GetProducts(page, limit int, search, category string) (*models.PaginatedProductResponse, error){
	// 1. Fallback to safe defaults if inputs are weird
	if page < 1{
		page = 1
	}
	if limit < 1 || limit > 100{
		limit = 10 // Max 100 items per page to protect the server
	}

	offset := (page - 1) * limit

	// 2. Run queries concurrently or sequentially
	totalItems, err := s.repo.CountProducts(search, category)
	if err != nil{
		return nil, err
	}

	products, err := s.repo.GetProducts(limit, offset, search, category)
	if err != nil{
		return nil, err
	}

	// 3. Prevent returning nil for empty arrays in JSON
	if products == nil{
		products = []models.Product{}
	}

	// 4. Calculate total pages (Ceiling division)
	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))

	// 5. Build response
	res := &models.PaginatedProductResponse{
		Data: products,
		Meta: models.PaginationMeta{
			CurrentPage: page,
			PageSize: limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}
	return res, nil
}

func (s *productService) GetProductDetail(productID string) (*models.Product, error){
	// Fetch the product
	product, err := s.repo.GetProductByID(productID)
	if err != nil{
		return nil, err
	}

	// Hide deactivated products from the public
	if !product.IsActive{
		slog.Warn("Attempted to view deactivated product", slog.String("product_id", productID))
		return nil, errors.New("Product not found")
	}

	return product, nil
}

func (s *productService) ValidateStock(req models.StockCheckRequest)([]models.StockCheckResult, bool, error){
	var results []models.StockCheckResult
	allAvailable := true

	for _, item := range req.Items{
		slog.Warn("Validating item", slog.String("received_id", item.ProductID))
		result := models.StockCheckResult{
			ProductID: item.ProductID,
			HasStock: false,
		}

		//Fetch the product
		product, err := s.repo.GetProductByID(item.ProductID)
		if err != nil{
			result.Message = "Product not found"
			allAvailable = false
			results = append(results, result)
			continue
		}

		//Check if it was deactivated 
		if !product.IsActive{
			result.Message = "Product is no longer available"
			allAvailable = false
			results = append(results, result)
			continue
		}

		//Check the actual stock quantity
		result.CurrentStock = product.Stock
		result.Price = product.Price
		result.SellerID = product.SellerID
		
		if product.Stock < item.RequestedQuantity{
			result.Message = "Insufficient stock"
			allAvailable = false
		}else {
			result.HasStock = true
			result.Message = "Stock available"
		}

		results = append(results, result)
	}
	return results, allAvailable, nil
}