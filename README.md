## E-Commerce sample project ##

Playing around with GO nothing special

run each service using `go run main.go`

test with these curl calls

`
curl -X POST http://localhost:8080/orders \
-H "Content-Type: application/json" \
-d '{"order_id": 1, "product_id": 101, "user_id": 1, "quantity": 5}'
`

`
curl -X POST http://localhost:8082/notify \
-H "Content-Type: application/json" \
-d '{"user_id": 1, "message": "This is a test notification."}'
`

`
curl -X POST http://localhost:8081/check-stock \
-H "Content-Type: application/json" \
-d '{"product_id": 101, "quantity": 5}'
`