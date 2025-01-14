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
su - postgres -c "psql notification_db -c \"CREATE TABLE notifications (
    notification_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    message VARCHAR(255) NOT NULL
);\""
su - postgres -c "psql notification_db -c \"ALTER TABLE notification_db OWNER TO notification_user;\""
su - postgres -c "psql notification_db -c \"INSERT INTO notifications (user_id, message) VALUES (1, 'Order Confirmed'), (2, 'Order Cancelled'), (3, 'Order on the way');\""