package handler

import (
	"context"
	"log/slog"
	"net/http"

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