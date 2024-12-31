package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type StockRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type StockResponse struct {
	ProductID   int  `json:"product_id"`
	IsAvailable bool `json:"is_available"`
}

var stock = map[int]int{
	101: 10,
	102: 5,
	103: 0,
}

func CheckStock(productID, quantity int) bool {
	available, exists := stock[productID]
	return exists && available >= quantity
}

func connectToRabbitMQ() *amqp.Connection {
	var conn *amqp.Connection
	var err error
	for retries := 0; retries < 5; retries++ {
		conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err == nil {
			return conn
		}
		log.Printf("Retrying RabbitMQ connection: attempt %d", retries+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	return nil
}

func main() {
	conn := connectToRabbitMQ()
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare queue
	_, err = ch.QueueDeclare("check_stock", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare check_stock queue: %v", err)
	}

	// Consume messages from the check_stock queue
	msgs, err := ch.Consume("check_stock", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Inventory Service waiting for messages...")

	for msg := range msgs {
		var req StockRequest
		err := json.Unmarshal(msg.Body, &req)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		response := StockResponse{
			ProductID:   req.ProductID,
			IsAvailable: CheckStock(req.ProductID, req.Quantity),
		}

		responseBody, _ := json.Marshal(response)
		err = ch.Publish("", msg.ReplyTo, false, false, amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: msg.CorrelationId,
			Body:          responseBody,
		})
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		}
	}
}
