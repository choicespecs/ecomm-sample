FROM docker.io/library/golang:1.20

# Install ping utility
RUN apt-get update && apt-get install -y iputils-ping && apt-get clean


WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o main .

CMD ["/app/main"]
