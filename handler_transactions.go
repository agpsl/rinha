package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/agpsl/rinha/database"
	"github.com/go-chi/chi/v5"
)

type requestDataStruct struct {
  Valor int   `json:"valor"`
  Tipo  string  `json:"tipo"`
  Desc  string  `json:"descricao"`
}

type responseDataStruct struct {
  Limite  int32 `json:"limite"`
  Saldo   int32 `json:"saldo"`
}

// Validate transaction value
func valTransactionValue (valor int) bool {
  if valor < 0 {
    return false
  }
  if fmt.Sprintf("%T", valor) != "int" {
    return false
  }
  return true
}

// Validate transaction type
func valTransactionType (tipo string) bool {
  switch tipo { 
    case 
      "d",
      "c":
      return true
  } 
  return false
}

// Validate transaction description
func valTransactionDesc (desc string) bool {
  if desc == "" {
    return false
  }
  if desc == "null" {
    return false
  }
  if len(desc) > 10 {
    return false
  }
  return true
}

// proccess transaction in a way to make sure we only have the correct data on the database
func processTransaction(ctx context.Context, db *sql.DB, q *database.Queries, cData database.UpdateCustomerParams, tData database.InsertTransactionParams) (updCust []database.Cliente, e error) {
  tx, err := db.Begin()
  if err != nil {
    return nil, err
  }
  defer tx.Rollback()
  qtx := q.WithTx(tx)

  updCust, cerr := qtx.UpdateCustomer(ctx, cData)
  if cerr != nil {
    return nil, cerr
  }

  terr := qtx.InsertTransaction(ctx, tData)
  if terr != nil {
    return nil, terr
  }

  return updCust, tx.Commit()

}

func (apiCfg *apiConfig) handleTransaction(w http.ResponseWriter, r *http.Request) {
  // Get customer ID from URL
  customer_id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
  if err != nil {
    respondWithJSON(w, 422, fmt.Sprintf("Can't parse customer id: %v", err))
    return
  }

  // Parse the transaction data
  transaction := requestDataStruct{}
  terr := json.NewDecoder(r.Body).Decode(&transaction)
  if terr != nil {
    respondWithError(w, 422, fmt.Sprintf("Couldn't parse request body: %v", err))
    return
  } 
  
  // Some validations
  if !valTransactionValue(transaction.Valor) {
    respondWithError(w, 422, "Invalid value")
    return
  }
  if !valTransactionType(transaction.Tipo) {
    respondWithError(w, 422, "Invalid transaction typ")
    return
  }
  if !valTransactionDesc(transaction.Desc) {
    respondWithError(w, 422, "Invalid description")
    return
  }

  // Get customer info
  customer, err := apiCfg.Queries.GetCustomer(r.Context(), int32(customer_id))
  if err != nil {
    respondWithError(w, 404, fmt.Sprintf("Customer not found: %v", err))
    return
  }
  
  // Value to add/subtract to/from account
  addToBalance := 0
  if transaction.Tipo == "c" {
    addToBalance = transaction.Valor
  } else if transaction.Tipo == "d" {
    if (customer.Limite * -1) > (customer.Saldo - int32(transaction.Valor)){
      respondWithError(w, 422, "Insufficient funds")
      return
    }
    addToBalance = transaction.Valor * -1
  }

  // Prepare info
  customerData := database.UpdateCustomerParams{
    ID: int32(customer_id),
    Saldo: int32(addToBalance),
  }
  transactionData := database.InsertTransactionParams{
    ClienteID: int32(customer_id),
    Valor: int32(transaction.Valor),
    Tipo: transaction.Tipo,
    Descricao: transaction.Desc,
  }

  // Try to create/update the rows
  updatedCustomer, err := processTransaction(r.Context(), apiCfg.Database, apiCfg.Queries, customerData, transactionData)
  if err != nil {
    respondWithError(w, 422, fmt.Sprintf("Could not process transaction: %v", err))
    return
  }

  // Everything went ok, respond with 200 and the customer info
  respondWithJSON(w, 200, responseDataStruct{
    Limite: updatedCustomer[0].Limite,
    Saldo: updatedCustomer[0].Saldo,
  })
} 
