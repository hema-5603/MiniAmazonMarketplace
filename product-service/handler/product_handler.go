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