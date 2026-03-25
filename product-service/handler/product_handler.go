package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

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
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 1. Extract the user info from the JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 2. Role-Based Access Control
	if role != "SELLER" && role != "ADMIN"{
		slog.Warn("Unauthorized product creation attempt", slog.String("request_id", reqID), slog.String("user_id", sellerID), slog.String("role",role))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success":false,
			"message":"Access denied: Only seller can create the products",
		})
	}

	// 3. Parse JSON payload
	var req models.CreateProductRequest
	if err := c.Bind(&req); err!=nil{
		slog.Warn("Invalid product payload", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	// 4.Pass to service
	product, err := h.service.CreateProduct(ctx,sellerID, req)
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
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 1. Get the Product ID from the URL(/api/v1/products/:id)
	productID := c.Param("id")

	// 2. Extract user info from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 3. Role-Based Access Control check
	if role != "SELLER" && role != "ADMIN"{
		slog.Warn("Unauthorized product updation attempt", slog.String("request_id", reqID), slog.String("user_id", sellerID), slog.String("role",role))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success" : false,
			"message" : "Only sellers can update the products",
		})
	}

	// 4. Bind Payload
	var req models.UpdateProductRequest
	if err := c.Bind(&req); err != nil{
		slog.Warn("Invalid product update payload", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success" : false,
			"message" : "Invalid request payload",
		})
	}

	// 5. Call Service
	updatedProduct, err := h.service.UpdateProduct(ctx, productID, sellerID, req)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Product not found"{
			status = http.StatusNotFound
		}else if err.Error() == "Unauthorized: You do not own this product"{
			status = http.StatusForbidden
		}else if err.Error() == "Invalid product data: Name and positive price are required"{
			status = http.StatusBadRequest
		}

		if status == http.StatusInternalServerError{
			slog.Error("Failed to update product", slog.String("request_id", reqID), slog.String("product_id",productID), slog.String("error",err.Error()))
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
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)
	
	productID := c.Param("id")

	// 1. Extract user info from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 2. Role-Based Access Control
	if role != "SELLER" && role != "ADMIN"{
		slog.Warn("Unauthorized product stock updation attempt", slog.String("user_id", sellerID), slog.String("role",role))
		return c.JSON(http.StatusForbidden,map[string]interface{}{
			"success" : false,
			"message" : "Access denied: Only sellers can update the product stock",
		})
	}

	// 3. Bind payload
	var req models.UpdateProductStockRequest
	if err := c.Bind(&req); err != nil{
		slog.Warn("Invalid product payload", slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success" : false,
			"message" : "Invalid request payload",
		})
	}

	// 4. Call service
	err := h.service.UpdateStock(ctx, productID, sellerID, req)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Product not found"{
			status = http.StatusNotFound
		}else if err.Error() == "Unauthorized: You do not own this product"{
			status = http.StatusForbidden
		}else if err.Error() == "Invalid operation: Stock cannot be negative"{
			status = http.StatusBadRequest
		}
		if status == http.StatusInternalServerError{
			slog.Error("Failed to update stock", slog.String("product_id",productID), slog.String("error",err.Error()))
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

func (h *ProductHandler) UpdateProductStatus(c echo.Context)error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	productID := c.Param("id")

	// 1. Extract user info from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	sellerID := claims["user_id"].(string)
	role := claims["role"].(string)

	// 2. Role-Based Access Check
	if role != "SELLER" && role != "ADMIN"{
		slog.Warn("Unauthorized product status updation attempt", slog.String("request_id", reqID), slog.String("user_id", sellerID), slog.String("role",role))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"message": "Access denied: Only sellers can change product status",
		})
	}

	// 3. Bind payload
	var req models.UpdateStatusRequest
	if err := c.Bind(&req); err != nil{
		slog.Warn("Invalid product status payload",slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest,map[string]interface{}{
			"success":false,
			"message": "Invalid request payload",
		})
	}

	// 4. Call service
	err := h.service.UpdateProductStatus(ctx, productID, sellerID, req)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Product not found"{
			status = http.StatusNotFound
		}else if err.Error() == "Unauthorized: You do not own this product"{
			status = http.StatusForbidden
		}

		if status == http.StatusInternalServerError{
			slog.Error("Failed to update status", slog.String("request_id", reqID), slog.String("product_id",productID), slog.String("error",err.Error()))
		}

		return c.JSON(status, map[string]interface{}{
			"success" : false,
			"message" : err.Error(),
		})
	}

	statusMsg := "Product deactivated successfully"
	if req.IsActive{
		statusMsg = "Product activated successfully"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success" : true,
		"message" : statusMsg,
	})
}

func (h *ProductHandler) GetProducts(c echo.Context) error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 1. Extract query paramaters
	search := c.QueryParam("search")
	category := c.QueryParam("category")

	//Convert strings to integers with defaults
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0{
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0{
		limit = 10
	}

	// 2. Call service
	response, err := h.service.GetProducts(ctx, page, limit, search, category)
	if err != nil{
		slog.Error("Failed to fetch products", slog.String("request_id", reqID), slog.String("error",err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to fetch products",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": response.Data,
		"meta": response.Meta,
	})
}

func (h *ProductHandler) GetProductDetail(c echo.Context) error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 1. Get ID from the URL (/api/v1/products/:id)
	productID := c.Param("id")
	
	// 2. Call service
	product, err := h.service.GetProductDetail(ctx, productID)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Product not found"{
			status = http.StatusNotFound
		}else{
			slog.Error("Failed to fetch product detail", slog.String("request_id", reqID), slog.String("error", err.Error()))
		}

		return c.JSON(status, map[string]interface{}{
			"success" : false,
			"message" : err.Error(),
		})
	}
	// 3. Return the product
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success" : true,
		"data" : product,
	})
}

func (h *ProductHandler) ValidateStock(c echo.Context) error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	var req models.StockCheckRequest

	// 1.Bind payload
	if err := c.Bind(&req); err != nil || len(req.Items) == 0{
		slog.Warn("Invalid product payload or empty items array", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":false,
			"message":"Invalid request payload or empty items array",
		})
	}

	// 2.Call service
	results, allAvailable, err := h.service.ValidateStock(ctx, req)
	if err != nil{
		slog.Error("Error durings stock validation", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success":false,
			"message":"An error occurred while validating stock",
		})
	}

	// 3. Returning the detailed report
		// Return 200 OK ,even if the stock is insufficient, because the request itself was successful
		//"allAvailable" boolean tells whether we can proceed
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":true,
			"all_available":allAvailable,
			"data":results,
		})
}