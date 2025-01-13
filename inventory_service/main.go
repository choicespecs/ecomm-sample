package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rabbitmq/amqp091-go"
)

type StockRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type StockResponse struct {
	ProductID   int  `json:"product_id"`
	IsAvailable bool `json:"is_available"`
}

type HealthResponse struct {
	Service  string `json:"service"`
	Status   string `json:"status"`
	Database string `json:"database"`
	Error    string `json:"error,omitempty"`
}

var db *sql.DB

func connectToDatabase() {
	var err error
	connStr := "postgres://inventory_user:password@localhost:5432/inventory_db?sslmode=disable"
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

func CheckStock(productID, quantity int) bool {
	var available int
	err := db.QueryRow("SELECT stock FROM inventory WHERE product_id = $1", productID).Scan(&available)
	if err != nil {
		if err == sql.ErrNoRows {
			return false // Product not found
		}
		log.Printf("Error checking stock for product ID %d: %v", productID, err)
		return false
	}
	return available >= quantity
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

func processCheckStockQueue(ch *amqp091.Channel) {
	msgs, err := ch.Consume("check_stock", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume check_stock queue: %v", err)
	}

	log.Println("Processing check_stock queue...")

	for msg := range msgs {
		var req StockRequest
		err := json.Unmarshal(msg.Body, &req)
		if err != nil {
			log.Printf("Failed to parse stock request: %v", err)
			continue
		}

		response := StockResponse{
			ProductID:   req.ProductID,
			IsAvailable: CheckStock(req.ProductID, req.Quantity),
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
			log.Printf("Failed to publish stock response: %v", err)
		}
	}
}

func listenForHealthCheck(ch *amqp091.Channel) {
    // Declare a unique, auto-deleted queue for this service
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

    // Bind the queue to the fanout exchange
    err = ch.QueueBind(
        queue.Name,               // Queue name
        "",                       // Routing key (ignored for fanout exchange)
        "health_check_exchange",  // Exchange name
        false,
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to bind queue to exchange: %v", err)
    }

    // Consume messages from the dynamically generated queue
    msgs, err := ch.Consume(
        queue.Name, // Consume from the unique queue
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

    log.Println("Stock Service listening for health_check requests...")

    for msg := range msgs {
        log.Printf("Received health check request: %s", msg.CorrelationId)

        dbStatus := "connected"
        if err := db.Ping(); err != nil {
            dbStatus = "disconnected"
        }

        response := HealthResponse{
            Service:  "Stock Service",
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
	// Connect to the database
	connectToDatabase()
	defer db.Close()

	// Connect to RabbitMQ
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

	// Declare the fanout exchange
	err = ch.ExchangeDeclare(
		"health_check_exchange", // Exchange name
		"fanout",                // Type
		true,                    // Durable
		false,                   // Auto-deleted
		false,                   // Internal
		false,                   // No-wait
		nil,                     // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare fanout exchange: %v", err)
	}

	go processCheckStockQueue(ch)
	go listenForHealthCheck(ch)

	select {}
}