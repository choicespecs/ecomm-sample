package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type OrderRequest struct {
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	UserID    int `json:"user_id"`
	Quantity  int `json:"quantity"`
}

type StockResponse struct {
	ProductID   int  `json:"product_id"`
	IsAvailable bool `json:"is_available"`
}

type Notification struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

var rabbitConn *amqp091.Connection
var rabbitChannel *amqp091.Channel

func connectToRabbitMQ() {
	var err error
	rabbitConn, err = amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	log.Println("Connected to RabbitMQ")
}

func unifiedHandler(w http.ResponseWriter, r *http.Request) {
	var orderReq OrderRequest
	err := json.NewDecoder(r.Body).Decode(&orderReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Step 1: Check Stock
	responseQueue, err := rabbitChannel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		http.Error(w, "Failed to declare response queue", http.StatusInternalServerError)
		return
	}

	corrID := "unified-correlation-id"
	requestBody, _ := json.Marshal(map[string]interface{}{
		"product_id": orderReq.ProductID,
		"quantity":   orderReq.Quantity,
	})

	err = rabbitChannel.PublishWithContext(
		nil,
		"",
		"check_stock",
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       responseQueue.Name,
			Body:          requestBody,
		},
	)
	if err != nil {
		http.Error(w, "Failed to publish stock-check message", http.StatusInternalServerError)
		return
	}

	msgs, err := rabbitChannel.Consume(responseQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		http.Error(w, "Failed to consume response queue", http.StatusInternalServerError)
		return
	}

	var stockResp StockResponse
	select {
	case msg := <-msgs:
		if msg.CorrelationId == corrID {
			err := json.Unmarshal(msg.Body, &stockResp)
			if err != nil {
				http.Error(w, "Invalid stock response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Correlation ID mismatch", http.StatusInternalServerError)
			return
		}
	case <-time.After(10 * time.Second):
		http.Error(w, "Timeout waiting for stock response", http.StatusGatewayTimeout)
		return
	}

	if !stockResp.IsAvailable {
		http.Error(w, "Stock not available", http.StatusConflict)
		return
	}

	// Step 2: Place Order
	orderBody, _ := json.Marshal(orderReq)
	err = rabbitChannel.PublishWithContext(
		nil,
		"",
		"place_order",
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        orderBody,
		},
	)
	if err != nil {
		http.Error(w, "Failed to publish order", http.StatusInternalServerError)
		return
	}

	// Step 3: Notify User
	notification := Notification{
		UserID:  orderReq.UserID,
		Message: "Your order has been successfully placed!",
	}
	notifyBody, _ := json.Marshal(notification)
	err = rabbitChannel.PublishWithContext(
		nil,
		"",
		"notifications",
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        notifyBody,
		},
	)
	if err != nil {
		http.Error(w, "Failed to publish notification", http.StatusInternalServerError)
		return
	}

	// Final Response
	response := map[string]interface{}{
		"order_id":    orderReq.OrderID,
		"status":      "Order processed successfully",
		"stock_check": stockResp,
		"notification": map[string]interface{}{
			"user_id":  notification.UserID,
			"message":  notification.Message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	connectToRabbitMQ()
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	http.HandleFunc("/api/process-order", unifiedHandler)

	log.Println("API Gateway running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
