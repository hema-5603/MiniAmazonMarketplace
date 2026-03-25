package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"product-service/mocks"
	"product-service/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//Helper to fake the JWT middleware output
func createMockUser(userID, role string) *jwt.Token{
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["role"] = role
	return token
}

func TestProductHandler(t *testing.T){
	e := echo.New()

	t.Run("CreateProduct - Failure(Customer Role)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		//Inject CUSTOMER role
		c.Set("user", createMockUser("cust1", "CUSTOMER"))

		_ = h.CreateProduct(c)

		//Assert HTTP 403 Forbidden
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("CreateProduct - Invalid payload (Type mismatch)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		badJSON := []byte(`{"name": "Laptop", "price": "two hundred"}`)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(badJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user",createMockUser("seller1", "SELLER"))

		_ = h.CreateProduct(c)

		//Assert HTTP 400 Bad request (Echo's bind will fail)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("CreateProduct - Internal server error(Database down)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		reqBody := models.CreateProductRequest{Name: "Blanket", Price: 500}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c:=e.NewContext(req, rec)

		// Inject SELLER role
		c.Set("user", createMockUser("seller1", "SELLER"))

		// Force the mock to simulate the database crash
		mockSvc.On("CreateProduct", mock.Anything, "seller1", reqBody).
			Return(nil, errors.New("Database connection lost")).Once()

		_ = h.CreateProduct(c)

		// Assert HTTP 500 Internal server error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("UpdateProduct - Unauthorized Role (CUSTOMER)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		body, _ := json.Marshal(models.UpdateProductRequest{Name: "Hi-end Laptop", Price: 1000.0})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/products/prod1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("prod1")

		//Inject a JWT with CUSTOMER role
		c.Set("user",createMockUser("cust1", "CUSTOMER"))

		_ = h.UpdateProduct(c)

		// Assert HTTP 403 Forbidden
		assert.Equal(t, http.StatusForbidden, rec.Code)

		// Ensure the mock service was never called since the handler blocked it early
		mockSvc.AssertNotCalled(t, "UpdateProduct")
	})

	t.Run("UpdateProduct - Invalid request payload(Bad JSON)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		badJSON := []byte(`{"name": "Laptop", "price": "expensive"}`)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/products/prod1", bytes.NewReader(badJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("prod1")

		c.Set("user",createMockUser("seller1", "SELLER"))

		_ = h.UpdateProduct(c)

		// Assert HTTP 400 Bad request
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// Ensure the mock service was never called since the handler blocked it early
		mockSvc.AssertNotCalled(t, "UpdateProduct")
	})

	t.Run("UpdateProduct - Internal server error (Database crash)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		reqBody := models.UpdateProductRequest{Name: "Blanket", Price: 500}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/products/prod1", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c:=e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues("prod1")

		// Inject SELLER role
		c.Set("user", createMockUser("seller1", "SELLER"))

		// Force the mock to simulate the database crash
		mockSvc.On("UpdateProduct", mock.Anything,"prod1", "seller1", reqBody).
			Return(nil, errors.New("Unexpected database timeout")).Once()

		_ = h.UpdateProduct(c)

		// Assert HTTP 500 Internal server error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
	t.Run("UpdateStock - Success", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		body, _ := json.Marshal(models.UpdateProductStockRequest{Stock: 50})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/prod1/stock", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("prod1")

		//Inject SELLER role
		c.Set("user",createMockUser("seller1", "SELLER"))

		mockSvc.On("UpdateStock", mock.Anything, "prod1", "seller1", mock.AnythingOfType("models.UpdateProductStockRequest")).Return(nil).Once()

		_ = h.UpdateStock(c)

		//Assert HTTP 200 OK
		assert.Equal(t, http.StatusOK, rec.Code)
		mockSvc.AssertExpectations(t)
	})
	t.Run("UpdateProductStatus - Failure(Not found)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		body, _ := json.Marshal(models.UpdateStatusRequest{IsActive: false})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/prod1/status", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("prod1")
		c.Set("user",createMockUser("seller1", "SELLER"))

		mockSvc.On("UpdateProductStatus", mock.Anything, "prod1", "seller1", mock.Anything).Return(errors.New("Product not found")).Once()

		_ = h.UpdateProductStatus(c)

		//Assert HTTP 404 NOT FOUND
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("UpdateProductStatus - Unauthorized role (Customer)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		body, _ := json.Marshal(models.UpdateStatusRequest{IsActive: false})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/prod1/status", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		//Inject CUSTOMER role
		c.Set("user", createMockUser("cust1", "CUSTOMER"))

		_ = h.UpdateProductStatus(c)

		// Assert 403 Forbidden
		assert.Equal(t, http.StatusForbidden, rec.Code)
		mockSvc.AssertNotCalled(t, "UpdateProductStatus")
	})

	t.Run("UpdateProductStatus - Internal server error (Database crash)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		body, _ := json.Marshal(models.UpdateStatusRequest{IsActive: false})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/prod1/status", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("prod1")
		c.Set("user",createMockUser("seller1", "SELLER"))

		// Force the mock to simulate the database crash
		mockSvc.On("UpdateProductStatus", mock.Anything,"prod1", "seller1", mock.Anything).
			Return(errors.New("Unexpected database error")).Once()

		_ = h.UpdateProductStatus(c)

		// Assert HTTP 500 Internal server error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
	t.Run("GetProducts - Success (Pagination & Public access)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)
		
		//No JWT is here since it's public route
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2&limit=5", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockResponse := &models.PaginatedProductResponse{
			Data: []models.Product{{ID: "1", Name: "Test"}},
			Meta: models.PaginationMeta{CurrentPage: 2, TotalPages: 3},
		}

		mockSvc.On("GetProducts", mock.Anything, 2, 5, "", "").Return(mockResponse, nil).Once()

		_ = h.GetProducts(c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GetProducts - Internal server error(Database down)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		c:=e.NewContext(req, rec)

		// Force the mock to simulate the database crash
		mockSvc.On("GetProducts", mock.Anything, 1, 10, "", "").
			Return(nil, errors.New("Database connection lost")).Once()

		_ = h.GetProducts(c)

		// Assert HTTP 500 Internal server error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("GetProductDetail - Internal server error(Database down)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/prod1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues("prod1")

		// Force the mock to simulate the database crash
		mockSvc.On("GetProductDetail", mock.Anything, "prod1").
			Return(nil, errors.New("Connection timeout")).Once()

		_ = h.GetProductDetail(c)

		// Assert HTTP 500 Internal server error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("ValidateStock - Failure(Invalid JSON)", func(t *testing.T) {
		mockSvc := new(mocks.MockProductService)
		h := NewProductHandler(mockSvc)

		// Sending broken JSON 
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products/validate-stock", bytes.NewReader([]byte(`{invalid json}`)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c:=e.NewContext(req, rec)

		_ = h.ValidateStock(c)

		// Assert HTTP 400 Bad request
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})	
}