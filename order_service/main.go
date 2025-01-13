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
func processOrderQueue(conn *amqp091.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
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

	msgs, err := ch.Consume(
		"response_order_service",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to consume from response queue: %v", err)
	}

	log.Println("Order Service waiting for stock responses...")

	for msg := range msgs {
		log.Printf("Stock response received: %s", msg.Body)

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

// listenForHealthCheck listens for health-check requests and responds.
func listenForHealthCheck(conn *amqp091.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"",    // Auto-generate queue name
		false, // Durable
		true,  // Auto-delete
		true,  // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		"",
		"health_check_exchange",
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue to exchange: %v", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to consume health_check queue: %v", err)
	}

	log.Println("Order Service listening for health_check requests...")

	for msg := range msgs {
		dbStatus := "connected"
		if err := db.Ping(); err != nil {
			dbStatus = "disconnected"
		}

		response := HealthResponse{
			Service:  "Order Service",
			Status:   "healthy",
			Database: dbStatus,
		}
		if dbStatus == "disconnected" {
			response.Status = "unhealthy"
			response.Error = "Database connection failed"
		}

		responseBody, _ := json.Marshal(response)
		err := ch.PublishWithContext(
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
		} else {
			log.Printf("Health response published: %v", response)
		}
	}
}

func main() {
	connectToDatabase()
	defer db.Close()

	conn := connectToRabbitMQ()
	defer conn.Close()

	err := func() error {
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		defer ch.Close()

		return ch.ExchangeDeclare(
			"health_check_exchange",
			"fanout",
			true,
			false,
			false,
			false,
			nil,
		)
	}()
	if err != nil {
		log.Fatalf("Failed to declare fanout exchange: %v", err)
	}

	go processOrderQueue(conn)
	go listenForHealthCheck(conn)

	select {}
}