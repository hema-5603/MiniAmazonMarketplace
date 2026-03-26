package main

import (
	"context"
	"log"
	"log/slog"
	"order-service/client"
	"order-service/config"
	"order-service/handler"
	"order-service/repository"
	"order-service/service"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout,nil))
	slog.SetDefault(logger)

	slog.Info("")

	//1. Load configurations
	cfg := config.LoadConfig()

	//2. Connect to Database
	db := config.ConnectDB(cfg)
	defer db.Close()

	//3. Initialize echo
	e := echo.New()
	e.Use(middleware.RequestID())
	//Request Logging
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc{
		return func(c echo.Context)error{
			start := time.Now()

			err:= next(c)
			if err!= nil{
				c.Error(err)
			}
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			slog.Info("http request",
				slog.String("request_id", reqID),
				slog.String("method",c.Request().Method),
				slog.String("uri",c.Request().RequestURI),
				slog.Int("status",c.Response().Status),
				slog.Duration("latency",time.Since(start)),
			)
			return nil		
		}
	})
	e.Use(middleware.Recover())
	//4. Initialize layers
	// Assume db is your *sql.DB connection
	orderRepo := repository.NewOrderRepository(db)

	// Initialize the HTTP client
	productClient := client.NewProductClient(cfg.ProductServiceURL)

	// Pass the client into the service
	orderService := service.NewOrderService(orderRepo, productClient)
	orderHandler := handler.NewOrderHandler(orderService)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "Order service is running",
		})
	})
	//5. Register routes
	// Grouping the API version
	v1 := e.Group("/api/v1")
	
	//6. Register protected routes
	//6.1 Create the protected group for User routes
	protectedGroup := v1.Group("")

	//6.2 Apply the JWT middleware
	protectedGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.JWTSecret),
	}))

	//6.3 Protected route

	// Checkout endpoint
	protectedGroup.POST("/orders/checkout",orderHandler.Checkout)

	// Get order history 
	protectedGroup.GET("/orders", orderHandler.GetOrderHistory)

	// Get order detail
	protectedGroup.GET("/orders/:id", orderHandler.GetOrderDetail)


	// The Background cron job
	go func(){
		//Run this loop forever in the background
		ticker := time.NewTicker(1*time.Minute) // Check every 1 minute
		defer ticker.Stop()

		for range ticker.C{
			// Provide a background context for the job
			ctx := context.Background()

			// Call the service for expire unpaid orders
			err := orderService.ExpireUnpaidOrders(ctx)
			if err != nil{
				slog.Error("Cron job execution failed", slog.String("error", err.Error()))
			}
		}
	}()

	//7. Start the server on the dynamic port
	port := cfg.ServerPort
	if port == ""{
		port = "8082"
	}
	log.Printf("Starting server on port %s",port)
	e.Logger.Fatal(e.Start(":"+port))
}