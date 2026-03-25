package config

import (
	"database/sql"
	"fmt"
	"log"

	_"github.com/go-sql-driver/mysql"
)

//ConnectDB builds the connection string from the config and connects to MYSQL

func ConnectDB(cfg *Config) *sql.DB{
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",cfg.DBUser,cfg.DBPassword,cfg.DBHost,cfg.DBPort,cfg.DBName) //"user:pass@tcp(127.0.0.1:3306)/dbname"

	db,err := sql.Open("mysql",dsn) 

	if err != nil{
		log.Fatal("Failed to open DB connection",err)
	}

	//Checking the connection is alive
	if err := db.Ping(); err!=nil{
		log.Fatal("Failed to ping DB:",err)
	}
	log.Println("Successfully connected to the database")
	return db
}