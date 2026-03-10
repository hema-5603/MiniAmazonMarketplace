package main

import (
	"log"
	"user-service/config"
	"user-service/handler"
	"user-service/repository"
	"user-service/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
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
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	//5. Register public routes
	// Grouping the API version
	v1 := e.Group("/api/v1")
	v1.POST("/auth/register", userHandler.Register)
	
	//6. Start the server on the dynamic port
	port := cfg.ServerPort
	if port == ""{
		port = "8080"
	}
	log.Printf("Starting server on port %s",port)
	e.Logger.Fatal(e.Start(":"+port))
}