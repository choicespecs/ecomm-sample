package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handleNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var notif Notification
	err := json.NewDecoder(r.Body).Decode(&notif)
	if err != nil {
		http.Error(w, "Invalid notification data", http.StatusBadRequest)
		return
	}

	Send(notif)
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/notify", handleNotification)
	log.Println("Notification Service running on http://localhost:8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
