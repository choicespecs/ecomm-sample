package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rabbitmq/amqp091-go"
)

type Order struct {
	OrderID   int    `json:"order_id"`
	ProductID int    `json:"product_id"`
	UserID    int    `json:"user_id"`
	Quantity  int    `json:"quantity"`
}

var db *sql.DB

func connectToDatabase() {
	var err error
	connStr := "postgres://order_user:password@localhost:5432/order_db?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database connection not established: %v", err)
	}
	log.Println("Connected to the database successfully.")
}

func connectToRabbitMQ() *amqp091.Connection {
	var conn *amqp091.Connection
	var err error
	for retries := 0; retries < 5; retries++ {
		conn, err = amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
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
	connectToDatabase()
	defer db.Close()
	
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

	log.Println("Order Service waiting for stock responses...")

	for msg := range msgs {
		// Process each stock response
		log.Printf("Stock response received: %s", msg.Body)

		// Publish notification after processing stock response
		notification := map[string]interface{}{
			"user_id":  1, // Replace with actual user ID
			"message":  "Order processed successfully!",
		}
		notifyBody, _ := json.Marshal(notification)
		err := ch.PublishWithContext(
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
}