package main

import (
	"log/slog"
	"os"
	"time"
	"product-service/config"
	"product-service/handler"
	"product-service/repository"
	"product-service/service"

	_ "github.com/go-sql-driver/mysql"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	//Initializing the standard JSON logger
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
	
	// Request ID middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	//Request Logging
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc{
		return func(c echo.Context)error{
			start := time.Now()

			err:= next(c)
			if err!= nil{
				c.Error(err)
			}
			// Request ID in the http log
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			slog.Info("http request",
				slog.String("request_id",reqID),
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
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	//5. Register public routes
	// Grouping the API version
	v1 := e.Group("/api/v1")
	
	// PUBLIC ROUTES 
	publicGroup := v1.Group("")
	publicGroup.GET("/products", productHandler.GetProducts)
	
	publicGroup.GET("/products/:id",productHandler.GetProductDetail)
	//6. Register protected routes
	//6.1 Create the protected group for User routes
	protectedGroup := v1.Group("")

	//6.2 Apply the JWT middleware
	protectedGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.JWTSecret),
	}))

	//6.3 Protected route
	//Create a Product ("/api/v1/products")
	protectedGroup.POST("/products",productHandler.CreateProduct)

	// Update the entire product (/api/v1/products/:id)
	protectedGroup.PUT("/products/:id",productHandler.UpdateProduct)

	// Update the product stock (/api/v1/products/:id/stock)
	protectedGroup.PATCH("/products/:id/stock",productHandler.UpdateStock)

	//Update the product status
	protectedGroup.PATCH("/products/:id/status",productHandler.UpdateProductStatus)


	// Validate stock
	publicGroup.POST("/products/validate-stock",productHandler.ValidateStock)

	// Reserve stock
	publicGroup.POST("/products/reserve-stock",productHandler.ReserveStock)
	//7. Start the server on the dynamic port
	port := cfg.ServerPort
	if port == ""{
		port = "8081"
	}
	slog.Info("Starting product service",slog.String("port",port))
	e.Logger.Fatal(e.Start(":"+port))
}