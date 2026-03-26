package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"order-service/models"
	"order-service/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type OrderHandler struct{
	service service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler{
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Checkout(c echo.Context) error{
	// 1. Context tracing
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 2. Extract user info from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// 3.Bind JSON payload
	var req models.CheckoutRequest
	if err := c.Bind(&req); err != nil{
		slog.Warn("Invalid checkout payload", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	// 4. Call service
	order, err := h.service.CreateOrder(ctx, userID, req)
	if err != nil{
		slog.Error("Checkout failed", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success" : false,
			"message": err.Error(),
		})
	}

	// 5. Return success
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success" : true,
		"message" : "Order placed successfully", 
		"data" : order,
	})
}


func (h *OrderHandler) GetOrderDetail(c echo.Context) error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 1. Extract order ID from the URL (/api/v1/orders/:id)
	orderID := c.Param("id")

	// 2. Extract User ID from the JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// 3. Call the service
	order, err := h.service.GetOrderDetail(ctx, orderID, userID)
	if err != nil{
		status := http.StatusInternalServerError
		if err.Error() == "Order not found"{
			status = http.StatusNotFound
		}else if err.Error() == "Unauthorized: You do not own this order"{
			status = http.StatusForbidden
		}

		return c.JSON(status, map[string]interface{}{
			"success" : false,
			"message" : err.Error(),
		})
	}

	// 4. Return success
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success" : true,
		"data" : order,
 	})
}

func (h *OrderHandler) GetOrderHistory(c echo.Context) error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)

	// 1. Extract userID from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// 2. Parse query parameters
	status := c.QueryParam("status")

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 10
	}

	// 3. Call service
	response, err := h.service.GetOrderHistory(ctx, userID, page, limit, status)
	if err != nil{
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success" : false,
			"message" : err.Error(),
		})
	}

	// 4. Return data
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success" : true,
		"data" : response.Data,
		"meta" : response.Meta,
	})
}

func (h *OrderHandler) CancelOrder(c echo.Context) error{
	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	ctx := context.WithValue(c.Request().Context(), models.RequestIDKey, reqID)


	// 1. Get Order ID from the URL path
	orderID := c.Param("id")

	// 2. Extract userID from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// 3. Call the service
	err := h.service.CancelOrder(ctx, orderID, userID)
	if err != nil{
		status := http.StatusBadRequest
		if err.Error() == "Unauthorized: You do not own this product" {
			status = http.StatusForbidden
		}else if err.Error() == "Order not found"{
			status = http.StatusNotFound
		}

		return c.JSON(status, map[string]interface{}{
			"success": false,
			"message" : err.Error(),
		})
	}

	// 4. Return success
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success" : true,
		"message" : "Order has been successfully cancelled",
	})
}