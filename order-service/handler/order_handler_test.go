package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"order-service/mocks"
	"order-service/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper to fake the JWT middleware output
func createMockUser(userID, role string) *jwt.Token{
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["role"] = role
	return token
}

func TestOrderHandler(t *testing.T){
	e := echo.New()

	t.Run("Checkout - Success", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		reqBody := models.CheckoutRequest{
			Items: []models.CheckoutItem{
				{
					ProductID: "prod1",
					Quantity: 1,
					Price: 200,
				},
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user", createMockUser("user1", "CUSTOMER"))

		mockOrder := &models.Order{
			ID: "order1",
			TotalAmount: 200,
		}
		mockSvc.On("CreateOrder", mock.Anything, "user1", reqBody).Return(mockOrder, nil).Once()

		_ = h.Checkout(c)

		assert.Equal(t, http.StatusCreated, rec.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Checkout -Failure(Bad request)", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		badJSON := []byte(`{"items": "this should be an array"}`)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewReader(badJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user", createMockUser("user1", "CUSTOMER"))

		_ = h.Checkout(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		mockSvc.AssertNotCalled(t, "CreateOrder")
	})

	t.Run("GetOrderDetail - Success", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/order1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("order1")
		c.Set("user", createMockUser("user1", "CUSTOMER"))

		mockOrder := &models.Order{
			ID: "order1",
			UserID: "user1",
		}
		mockSvc.On("GetOrderDetail", mock.Anything, "order1", "user1").Return(mockOrder, nil).Once()

		_ = h.GetOrderDetail(c)
		
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("GetOrderDetail - Failure (404 Not Found)", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/order99", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("order99")
		c.Set("user", createMockUser("user1", "CUSTOMER"))

		mockSvc.On("GetOrderDetail", mock.Anything, "order99", "user1").Return(nil, errors.New("Order not found")).Once()

		_ = h.GetOrderDetail(c)
		
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
	t.Run("GetOrderDetail - Failure (Forbidden 403)", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/order1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("order1")
		c.Set("user", createMockUser("user2", "CUSTOMER"))

		mockSvc.On("GetOrderDetail", mock.Anything, "order1", "user2").Return(nil, errors.New("Unauthorized: You do not own this order")).Once()

		_ = h.GetOrderDetail(c)
		
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("GetOrderHistory - Success", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user", createMockUser("user1", "CUSTOMER"))

		mockResponse := &models.PaginatedOrderResponse{
			Data: []models.Order{{
				ID: "order1",
			}},
			Meta: models.PaginatedMeta{
					CurrentPage: 1,
					TotalPages: 1,
			},
		}	

		mockSvc.On("GetOrderHistory", mock.Anything,"user1", 1, 10, "").Return(mockResponse,nil).Once()

		_ = h.GetOrderHistory(c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("GetOrderHistory - Failure (Internal server error)", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user", createMockUser("user1", "CUSTOMER"))

		mockSvc.On("GetOrderHistory", mock.Anything,"user1", 1, 10, "").Return(nil, errors.New("Failed to retrieve order history")).Once()

		_ = h.GetOrderHistory(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
	t.Run("CancelOrder - Failure(Forbidden)", func(t *testing.T) {
		mockSvc := new(mocks.MockOrderService)
		h := NewOrderHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/orders/order1/cancel", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("order1")

		c.Set("user", createMockUser("user2", "CUSTOMER"))

		mockSvc.On("CancelOrder", mock.Anything, "order1", "user2").Return(errors.New("Unauthorized: You do not own this product")).Once()

		_ = h.CancelOrder(c)
		
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}