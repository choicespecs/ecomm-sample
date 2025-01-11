package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rabbitmq/amqp091-go"
)

type Notification struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

var db *sql.DB

func connectToDatabase() {
	var err error
	connStr := "postgres://notification_user:password@localhost:5432/notification_db?sslmode=disable"
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

	// Declare queue
	_, err = ch.QueueDeclare("notifications", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare notifications queue: %v", err)
	}

	// Consume messages from the notifications queue
	msgs, err := ch.Consume("notifications", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Notification Service waiting for messages...")

	for msg := range msgs {
		var notif Notification
		err := json.Unmarshal(msg.Body, &notif)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		log.Printf("Notification sent to UserID %d: %s", notif.UserID, notif.Message)
	}
}