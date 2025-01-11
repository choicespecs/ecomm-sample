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
su - postgres -c "psql -c \"CREATE USER notification_user WITH PASSWORD 'password';\""
su - postgres -c "psql -c \"CREATE DATABASE notification_db OWNER notification_user;\""
su - postgres -c "psql notification_db -c \"CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL
);\""
su - postgres -c "psql inventory_db -c \"ALTER TABLE notification_db OWNER TO notification_user;\""
su - postgres -c "psql notification_db -c \"INSERT INTO users (name, email) VALUES ('John Smith', 'john.smith@test.com'), ('David Barrow', 'david.man@example.com'), ('Diana Terry', 'dterr49@test.com');\""