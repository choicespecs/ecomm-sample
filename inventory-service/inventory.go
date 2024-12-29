package main

var stock = map[int]int{
	101: 10,
	102: 5,
	103: 0,
}

type StockRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func CheckStock(productID, quantity int) bool {
	available, exists := stock[productID]
	return exists && available >= quantity
}
