package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"database/sql"

	"github.com/agpsl/rinha/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
  Queries *database.Queries
  Database *sql.DB
}

func main() {
  PORT := fmt.Sprintf(":%s", os.Getenv("PORT"))
  DB_HOST := os.Getenv("DB_HOSTNAME")
  DB_NAME := "rinha"
  DB_USER := "admin"
  DB_PASSWORD := "123"
  DBSTRING := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
  
  // Database
  conn, err := sql.Open("postgres", DBSTRING)
  if err != nil {
    log.Printf("Can't connect to database: %+v", err)
    return
  }
 
  apiCfg := apiConfig{
    Queries: database.New(conn),
    Database: conn,
  }
  if err != nil {
    log.Printf("Can't create database connection: %+v", err)
  }

  // Routes
  http.HandleFunc("GET /status", handlerStatus)
  http.HandleFunc("GET /clientes/{id}/extrato", apiCfg.handleGetBalance)
  http.HandleFunc("POST /clientes/{id}/transacoes", apiCfg.handleTransaction)
  
  fmt.Println("Listening on", PORT)
  http.ListenAndServe(PORT, nil)
}
