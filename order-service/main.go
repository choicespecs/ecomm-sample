package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Order struct {
	OrderID   int    `json:"order_id"`
	ProductID int    `json:"product_id"`
	UserID    int    `json:"user_id"`
	Quantity  int    `json:"quantity"`
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Invalid order data", http.StatusBadRequest)
		return
	}

	// Call the Inventory Service to check stock
	stockAvailable, err := checkStock(order.ProductID, order.Quantity)
	if err != nil {
		http.Error(w, "Failed to check stock: "+err.Error(), http.StatusInternalServerError)
		sendNotification(order.UserID, "Order failed due to an internal error.")
		return
	}

	if !stockAvailable {
		http.Error(w, "Insufficient stock for product", http.StatusConflict)
		sendNotification(order.UserID, fmt.Sprintf("Order failed: insufficient stock for product %d.", order.ProductID))
		return
	}

	sendNotification(order.UserID, fmt.Sprintf("Order successfully placed for product %d.", order.ProductID))
	fmt.Fprintf(w, "Order accepted: %+v\n", order)
}

func checkStock(productID, quantity int) (bool, error) {
	requestBody, _ := json.Marshal(map[string]int{
		"product_id": productID,
		"quantity":   quantity,
	})

	resp, err := http.Post("http://localhost:8081/check-stock", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

func sendNotification(userID int, message string) {
	notification := map[string]interface{}{
		"user_id":  userID,
		"message":  message,
	}

	requestBody, _ := json.Marshal(notification)

	resp, err := http.Post("http://localhost:8082/notify", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Failed to send notification to UserID %d: %v", userID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Notification sent to UserID %d: %s", userID, message)
	} else {
		log.Printf("Failed to send notification to UserID %d, status: %d", userID, resp.StatusCode)
	}
}

func main() {
	http.HandleFunc("/orders", handleOrder)
	log.Println("Order Service running on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
