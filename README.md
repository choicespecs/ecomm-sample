## E-Commerce sample project ##

Playing around with GO nothing special

make sure rabbitmq is set up appropriately `podman-compose up` / `podman-compose down`
run each service using `go run main.go`

`curl -X POST -H "Content-Type: application/json" -d '{"order_id":1,"product_id":101,"user_id":1,"quantity":3}' http://localhost:8080/api/process-order`