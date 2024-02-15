package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type custInfoStruct struct {
  Balance int32     `json:"total"`
  Date    time.Time `json:"data_extrato"`
  Limit   int32     `json:"limite"`
}

type transStruct struct {
  Value int32      `json:"valor"`
  Type  string     `json:"tipo"`
  Desc  string     `json:"descricao"`
  When  time.Time  `json:"realizada_em"`
}

type custLastTransStruct struct {
  CustomerInfo custInfoStruct   `json:"saldo"`
  Transactions []transStruct    `json:"ultimas_transacoes"`
}

func (apiCfg *apiConfig) handleGetBalance(w http.ResponseWriter, r *http.Request) {
  // Get customer ID from URL
  customer_id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
  if err != nil {
    respondWithError(w, 400, fmt.Sprintf("Can't parse customer id: %v", err))
    return
  }
  
  // Get customer info
  customer, err := apiCfg.Queries.GetCustomer(r.Context(), int32(customer_id))
  if err != nil {
    respondWithError(w, 404, fmt.Sprintf("Customer not found: %v", err))
    return
  }
  
  // Get customer transactions
  customerTrans, err := apiCfg.Queries.GetLastTransactions(r.Context(), int32(customer_id))
  if err != nil {
    respondWithError(w, 404, fmt.Sprintf("Can't load customer transactions: %v", err))
    return
  }
  
  lastTransactions := []transStruct{}
  for _, v := range customerTrans {
    t := transStruct{
      Value: v.Valor,
      Type: v.Tipo,
      Desc: v.Descricao,
      When: v.RealizadaEm,
    }
    lastTransactions = append(lastTransactions, t)
  }

  // Response data
  customerData := custInfoStruct {
    Balance: customer.Saldo,
    Date: time.Now(),
    Limit: customer.Limite,
  }

  respondWithJSON(w, 200, custLastTransStruct{
    CustomerInfo: customerData,
    Transactions: lastTransactions,
  })
}
