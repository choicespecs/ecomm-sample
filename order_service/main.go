package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Order struct {
	OrderID   int    `json:"order_id"`
	ProductID int    `json:"product_id"`
	UserID    int    `json:"user_id"`
	Quantity  int    `json:"quantity"`
}

func connectToRabbitMQ() *amqp091.Connection {
	var conn *amqp091.Connection
	var err error
	for retries := 0; retries < 5; retries++ {
		conn, err = amqp091.Dial("amqp://guest:guest@localhost:5672/")
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

	// Declare queues
	_, err = ch.QueueDeclare("check_stock", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare check_stock queue: %v", err)
	}

	responseQueue, err := ch.QueueDeclare("response_order_service", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare response queue: %v", err)
	}

	_, err = ch.QueueDeclare("notifications", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare notifications queue: %v", err)
	}

	// Consume messages from the response queue
	msgs, err := ch.Consume(responseQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume from response queue: %v", err)
	}

	// Simulate order processing
	order := Order{OrderID: 1, ProductID: 101, UserID: 1, Quantity: 3}

	// Publish stock-check request
	corrID := "random-correlation-id"
	requestBody, _ := json.Marshal(order)
	err = ch.PublishWithContext(
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
		log.Fatalf("Failed to publish stock-check request: %v", err)
	}

	log.Println("Waiting for stock response...")
	for msg := range msgs {
		if msg.CorrelationId == corrID {
			// Process the response
			log.Printf("Stock response received: %s", msg.Body)
			break
		}
	}

	// Publish notification
	notification := map[string]interface{}{
		"user_id":  order.UserID,
		"message":  "Order processed successfully!",
	}
	notifyBody, _ := json.Marshal(notification)
	err = ch.PublishWithContext(
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
		log.Printf("Failed to publish notification: %v", err)
	}
}
