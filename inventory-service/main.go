package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handleCheckStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var request StockRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if CheckStock(request.ProductID, request.Quantity) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Insufficient stock", http.StatusConflict)
	}
}

func main() {
	http.HandleFunc("/check-stock", handleCheckStock)
	log.Println("Inventory Service running on http://localhost:8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
