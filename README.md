## E-Commerce sample project ##

Playing around with GO nothing special

Check RabbitMQ UI:

Open `http://localhost:15672` in a browser (username: guest, password: guest).
Verify that the expected queues (check_stock, notifications, etc.) are created.

`./startup` will set up docker containers, service.log, and start api gateway service

`./stop.sh` will tear down docker containers, delete service.logs, and shut down api gateway on :8080

`curl -X POST -H "Content-Type: application/json" -d '{"order_id":1,"product_id":101,"user_id":1,"quantity":3}' http://localhost:8080/api/process-order`

`curl -X GET http://localhost:8080/api/health-check`