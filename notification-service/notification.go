package main

import "log"

type Notification struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func Send(notif Notification) {
	log.Printf("Notification sent to UserID %d: %s", notif.UserID, notif.Message)
}
