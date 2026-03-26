package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"order-service/models"
)

// The response structure expected from the Product Service
type ProductValidationResponse struct{
	Success bool `json:"success"`
	AllAvailable bool `json:"all_available"`
	Data []struct{
		ProductID string `json:"product_id"`
		HasStock bool `json:"has_stock"`
		Price float64 `json:"price"`
		SellerID string `json:"seller_id"`
		Message string `json:"message"`
	} `json:"data"`
}

type ProductClient interface{
	ValidateCart(ctx context.Context, items []models.CheckoutItem) (*ProductValidationResponse, error)
	ReserveStock(ctx context.Context, items []models.CheckoutItem) error
	ReleaseStock(ctx context.Context, items []models.OrderItem) error
}

type productClient struct{
	baseURL string
	httpClient *http.Client
}

func NewProductClient(baseURL string) ProductClient{
	return &productClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // Never let the microservice hang forever
		},
	}
}

func (c *productClient) ValidateCart(ctx context.Context, items []models.CheckoutItem) (*ProductValidationResponse, error){
	reqID, _ := ctx.Value(models.RequestIDKey).(string)
	// 1. Prepare the payload
	payload := map[string]interface{}{
		"items" : items,
	}
	jsonData, _ := json.Marshal(payload)

	// 2. Create the HTTP request
	url := fmt.Sprintf("%s/api/v1/products/validate-stock", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil{
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	slog.Debug("Calling product service to validate cart", slog.String("request_id", reqID))

	// 3. Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil{
		slog.Error("Product service HTTP call failed", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return nil, errors.New("Product service is currently unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		slog.Warn("Product service return failure status", slog.String("request_id", reqID), slog.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("Product service returned status: %d", resp.StatusCode)
	}

	// 4. Decode the response
	var result ProductValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil{
		slog.Error("Failed to decode product service response", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return nil, errors.New("Failed to decode product service response")
	}

	return &result, nil
}

func (c *productClient) ReserveStock(ctx context.Context, items []models.CheckoutItem) error{
	reqID, _ := ctx.Value(models.RequestIDKey).(string)

	// Map the checkout items to the payload expected by the product service
	var reserveItems []map[string]interface{}
	for _, item := range items{
		reserveItems = append(reserveItems, map[string]interface{}{
			"product_id" : item.ProductID, 
			"quantity" : item.Quantity,
		})
	}

	payload := map[string]interface{}{
		"items" : reserveItems,
	}
	jsonData, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v1/products/reserve-stock", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil{
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	slog.Info("Calling product service to reserve stock", slog.String("request_id", reqID))

	resp, err := c.httpClient.Do(req)
	if err != nil{
		slog.Error("Product service HTTP call failed during reservation", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return errors.New("Product service is currently unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		// 409 conflict - When the inventory is empty
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		msg, _ := errorResponse["message"].(string)

		slog.Warn("Stock reservation rejected by product service", slog.String("request_id", reqID), slog.String("reason", msg))
		return errors.New(msg)
	}
	slog.Info("Stock reserved successfully", slog.String("request_id", reqID))
	return nil
}

func (c *productClient) ReleaseStock(ctx context.Context, items []models.OrderItem) error{
	var releaseItems []map[string]interface{}
	for _, items := range items{
		releaseItems = append(releaseItems, map[string]interface{}{
			"product_id": items.ProductID,
			"quantity" : items.Quantity,
		})
	}

	payload := map[string]interface{}{
		"items" : releaseItems,
	}

	jsonData, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v1/products/release-stock", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil{
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK{
		return errors.New("Failed to release stock in product service")
	}

	defer resp.Body.Close()
	return nil
}