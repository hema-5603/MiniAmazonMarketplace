package config

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//ConnectDB builds the connection string from the config and connects to MYSQL

func ConnectDB(cfg *Config) *sql.DB{
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",cfg.DBUser,cfg.DBPassword,cfg.DBHost,cfg.DBPort,cfg.DBName) //"user:pass@tcp(127.0.0.1:3306)/dbname"

	db,err := sql.Open("mysql",dsn) 

	if err != nil{
		slog.Error("Failed to open DB connection", slog.String("error", err.Error()))
		os.Exit(1)
	}

	//Checking the connection is alive
	if err := db.Ping(); err!=nil{
		slog.Error("Failed to ping DB:",slog.String("error",err.Error()))
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)   	//Maximum incoming connections
	db.SetMaxIdleConns(5) 		//Maximum idle connections kept alive
	db.SetConnMaxLifetime(5 * time.Minute) //How long a connection can live 
	slog.Info("Successfully connected to the database",slog.String("db_name",cfg.DBName))
	return db
}