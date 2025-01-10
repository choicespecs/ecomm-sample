#!/bin/bash

# Start PostgreSQL
service postgresql start

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to initialize..."
until pg_isready -h localhost -p 5432 -U postgres; do
  sleep 1
done
echo "PostgreSQL is ready!"

# Create user, database, and tables
su - postgres -c "psql -c \"CREATE USER order_user WITH PASSWORD 'password';\""
su - postgres -c "psql -c \"CREATE DATABASE order_db OWNER order_user;\""
su - postgres -c "psql -c \"GRANT ALL PRIVILEGES ON DATABASE order_db TO order_user;\""
su - postgres -c "psql order_db -c \"CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL,
    stock INT NOT NULL,
    name VARCHAR(100) NOT NULL
);\""

service postgresql restart