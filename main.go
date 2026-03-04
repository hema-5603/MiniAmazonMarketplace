package main

import(
	"database/sql"
	"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"

	"MAM/handler"
	"MAM/repository"
	"MAM/service"

)

func main() {
	//Connection to MySQL
	db,err := sql.Open("mysql","root:mam0103worlder@tcp(127.0.0.1:3306)/mam_user?parseTime=true") //"user:pass@tcp(127.0.0.1:3306)/dbname"

	if err != nil{
		log.Fatal("Failed to open DB connection",err)
	}
	defer db.Close()

	//Checking the connection is alive
	if err := db.Ping(); err!=nil{
		log.Fatal("Failed to ping DB:",err)
	}
	log.Println("Successfully connected to the database")
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