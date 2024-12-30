package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var stock = map[int]int{
	101: 10,
	102: 5,
	103: 0,
}

type StockRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func CheckStock(productID, quantity int) bool {
	available, exists := stock[productID]
	return exists && available >= quantity
}

func handleCheckStock(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received on /check-stock")

	if r.Method != http.MethodPost {
		log.Println("Invalid method:", r.Method)
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var request StockRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("Error decoding request:", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	log.Printf("Checking stock for ProductID: %d, Quantity: %d\n", request.ProductID, request.Quantity)
	if CheckStock(request.ProductID, request.Quantity) {
		log.Println("Stock is available")
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("Insufficient stock")
		http.Error(w, "Insufficient stock", http.StatusConflict)
	}
}

func main() {
	http.HandleFunc("/check-stock", handleCheckStock)
	log.Println("Inventory Service running on http://localhost:8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
