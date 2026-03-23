package main

import (
	"log"
	"log/slog"
	"os"
	"time"
	"user-service/config"
	"user-service/handler"
	"user-service/repository"
	"user-service/service"

	_ "github.com/go-sql-driver/mysql"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	// "github.com/labstack/echo/v4/middleware"
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
	
	//Request Logging
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc{
		return func(c echo.Context)error{
			start := time.Now()

			err:= next(c)
			if err!= nil{
				c.Error(err)
			}
			slog.Info("http request",
				slog.String("method",c.Request().Method),
				slog.String("uri",c.Request().RequestURI),
				slog.Int("status",c.Response().Status),
				slog.Duration("latency",time.Since(start)),
			)
			return nil		
		}
	})
	// e.Use(middleware.RequestLoggerWithConfig(middleware.LoggerConfig{
	// 	Format: `{
	// 	"time":"${time_rfc3339}",
	// 	"level":"INFO",
	// 	"prefix":"echo"
	// 	"method":"${method}",
	// 	"uri":"${uri}",
	// 	"status":${status},
	// 	"error":"${error}",
	// 	"latency_human":"${latency_human}"
	// 	}`, + "\n"
	// }))
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
	usersGroup.PUT("/profile",userHandler.UpdateProfile)

	//7. Start the server on the dynamic port
	port := cfg.ServerPort
	if port == ""{
		port = "8080"
	}
	log.Printf("Starting server on port %s",port)
	e.Logger.Fatal(e.Start(":"+port))
}