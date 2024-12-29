package main

type Order struct {
	OrderID   int    `json:"order_id"`
	ProductID int    `json:"product_id"`
	UserID    int    `json:"user_id"`
	Quantity  int    `json:"quantity"`
}
