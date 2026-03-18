package handler

import (
	"log/slog"
	"net/http"

	"product-service/models"
	"product-service/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type ProductHandler struct{
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler{
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) CreateProduct(c echo.Context) error{
	// 1. Extract the user info from the JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 2. Role-Based Access Control
	if role != "SELLER" && role != "ADMIN"{
		slog.Warn("Unauthorized product creation attempt", slog.String("user_id", sellerID), slog.String("role",role))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success":false,
			"message":"Access denied: Only seller can create the products",
		})
	}

	// 3. Parse JSON payload
	var req models.CreateProductRequest
	if err := c.Bind(&req); err!=nil{
		slog.Warn("Invalid product payload", slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":false,
			"message": "Invalid request payload",
		})
	}

	// 4.Pass to service
	product, err := h.service.CreateProduct(sellerID, req)
	if err != nil{
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success":false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success":true,
		"message": "Product create successfully",
		"data":product,
	})
}

func (h *ProductHandler) UpdateProduct(c echo.Context) error{
	// 1. Get the Product ID from the URL(/api/v1/products/:id)
	productID := c.Param("id")

	// 2. Extract user info from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 3. Role-Based Access Control check
	if role != "SELLER" && role != "ADMIN"{
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success" : false,
			"message" : "Only sellers can update the products",
		})
	}

	// 4. Bind Payload
	var req models.UpdateProductRequest
	if err := c.Bind(&req); err != nil{
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success" : false,
			"message" : "Invalid request payload",
		})
	}

	// 5. Call Service
	updatedProduct, err := h.service.UpdateProduct(productID, sellerID, req)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Product not found"{
			status = http.StatusNotFound
		}else if err.Error() == "Unauthorized: You do not own this product"{
			status = http.StatusForbidden
		}else if err.Error() == "Invalid product data: Name and positive price are required"{
			status = http.StatusBadRequest
		}

		return c.JSON(status, map[string]interface{}{
			"success" : false,
			"message" : err.Error(),
		})
	}

	return c.JSON(http.StatusOK,map[string]interface{}{
		"success" : true,
		"message": "Product updated successfully",
		"data" : updatedProduct,
	})
}

// Handler for updating the stock
func (h *ProductHandler) UpdateStock(c echo.Context) error{
	productID := c.Param("id")

	// 1. Extract user info from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 2. Role-Based Access Control
	if role != "SELLER" && role != "ADMIN"{
		return c.JSON(http.StatusForbidden,map[string]interface{}{
			"success" : false,
			"message" : "Access denied: Only sellers can update the product stock",
		})
	}

	// 3. Bind payload
	var req models.UpdateProductStockRequest
	if err := c.Bind(&req); err != nil{
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success" : false,
			"message" : "Invalid request payload",
		})
	}

	// 4. Call service
	err := h.service.UpdateStock(productID, sellerID, req)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Product not found"{
			status = http.StatusNotFound
		}else if err.Error() == "Unauthorized: You do not own this product"{
			status = http.StatusForbidden
		}else if err.Error() == "Invalid operation: Stock cannot be negative"{
			status = http.StatusBadRequest
		}

		return c.JSON(status, map[string]interface{}{
			"success" : false,
			"message" : err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success" : true,
		"message" : "Stock updated successfully",
	})
}