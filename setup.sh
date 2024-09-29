#!/bin/sh

for file in sql/*.sql; do
    if [ -f "$file" ]; then
        cat "$file" | psql -h localhost -d postgres -U postgres
    fi        
done
