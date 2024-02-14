package main

import (
	"net/http"

)

func handlerStatus(w http.ResponseWriter, r *http.Request) {
  respondWithJSON(w, 200, struct{}{})
}
