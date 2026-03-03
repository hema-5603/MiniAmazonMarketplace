package main

import(
	"database/sql",
	"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	
)

func main() {
	e := echo.New()

	// Assume db is your *sql.DB connection
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	// Grouping the API version
	v1 := e.Group("/api/v1")
	v1.POST("/auth/register", userHandler.Register)
	e.Logger.Fatal(e.Start(":8080"))
}