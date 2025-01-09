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
su - postgres -c "psql -c \"CREATE USER inventory_user WITH PASSWORD 'password';\""
su - postgres -c "psql -c \"CREATE DATABASE inventory_db OWNER inventory_user;\""
su - postgres -c "psql inventory_db -c \"CREATE TABLE inventory (
    product_id SERIAL PRIMARY KEY,
    stock INT NOT NULL
);\""
su - postgres -c "psql inventory_db -c \"INSERT INTO inventory (product_id, stock) VALUES (101, 10), (102, 5), (103, 0);\""