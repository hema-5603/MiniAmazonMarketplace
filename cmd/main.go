package main

import (
	"log"
	"user-service/config"
	"user-service/handler"
	"user-service/repository"
	"user-service/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
)

func main() {
	//1. Load configurations
	cfg := config.LoadConfig()

	//2. Connect to Database
	db := config.ConnectDB(cfg)
	defer db.Close()

	//3. Initialize echo
	e := echo.New()
	
	//4. Initialize layers
	// Assume db is your *sql.DB connection
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo,cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	//5. Register public routes
	// Grouping the API version
	v1 := e.Group("/api/v1")
	v1.POST("/auth/register", userHandler.Register)
	v1.POST("/auth/login",userHandler.Login)
	
	//6. Register protected routes
	//6.1 Create the protected group for User routes
	usersGroup := v1.Group("/users")

	//6.2 Apply the JWT middleware
	usersGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.JWTSecret),
	}))

	//6.3 Protected route
	usersGroup.GET("/profile",userHandler.GetProfile)

	//7. Start the server on the dynamic port
	port := cfg.ServerPort
	if port == ""{
		port = "8080"
	}
	log.Printf("Starting server on port %s",port)
	e.Logger.Fatal(e.Start(":"+port))
}