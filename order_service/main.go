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
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	UserID    int `json:"user_id"`
	Quantity  int `json:"quantity"`
}

type HealthResponse struct {
	Service  string `json:"service"`
	Status   string `json:"status"`
	Database string `json:"database"`
	Error    string `json:"error,omitempty"`
}

var db *sql.DB

// connectToDatabase establishes a connection to the PostgreSQL database.
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

// connectToRabbitMQ establishes a connection to RabbitMQ with retries.
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

// processOrderQueue listens for messages on the response_order_service queue.
func processOrderQueue(ch *amqp091.Channel) {
	// Declare the response queue
	_, err := ch.QueueDeclare(
		"response_order_service",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare response queue: %v", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		"response_order_service",
		"",    // consumer name
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to consume from response queue: %v", err)
	}

	log.Println("Order Service waiting for stock responses...")

	// Process each message
	for msg := range msgs {
		log.Printf("Stock response received: %s", msg.Body)

		// Publish notification after processing stock response
		notification := map[string]interface{}{
			"user_id":  1, // Replace with actual user ID
			"message":  "Order processed successfully!",
		}
		notifyBody, _ := json.Marshal(notification)
		err := ch.PublishWithContext(
			nil,
			"",               // exchange
			"notifications",  // routing key
			false,            // mandatory
			false,            // immediate
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

// listenForHealthCheck listens for health check requests on the health_check queue.
func listenForHealthCheck(ch *amqp091.Channel) {
	// Start consuming messages from the health_check queue
	msgs, err := ch.Consume(
		"health_check",
		"",    // consumer name
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to consume health_check queue: %v", err)
	}

	log.Println("Processing health_check queue...")

	// Process each health check message
	for msg := range msgs {
		err := db.Ping()
		dbStatus := "connected"
		if err != nil {
			dbStatus = "disconnected"
		}

		response := HealthResponse{
			Service:  "Order Service",
			Status:   "healthy",
			Database: dbStatus,
		}
		if dbStatus == "disconnected" {
			response.Status = "unhealthy"
			response.Error = err.Error()
		}

		responseBody, _ := json.Marshal(response)
		err = ch.PublishWithContext(
			nil,
			"",
			msg.ReplyTo,
			false,
			false,
			amqp091.Publishing{
				ContentType:   "application/json",
				CorrelationId: msg.CorrelationId,
				Body:          responseBody,
			},
		)
		if err != nil {
			log.Printf("Failed to publish health response: %v", err)
		}
	}
}

// main initializes the service and starts the goroutines for message handling.
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

	// Declare necessary queues
	queues := []string{"check_stock", "response_order_service", "notifications", "health_check"}
	for _, queue := range queues {
		_, err := ch.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", queue, err)
		}
	}

	// Start processing in separate goroutines
	go processOrderQueue(ch)
	go listenForHealthCheck(ch)

	// Prevent main from exiting
	select {}
}
