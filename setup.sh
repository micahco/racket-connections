#!/bin/sh

# Execute SQL queries
for file in sql/*.sql; do
    if [ -f "$file" ]; then
        cat "$file" | psql -h localhost -d postgres -U postgres
    fi        
done

# Create local developer env
cp .env.public .env
