#!/bin/bash
set -e

echo "Running database migrations..."

# Get all .up.sql files in order
for migration in $(ls migrations/*.up.sql | sort -V); do
    echo "Applying migration: $migration"
    go run ./cmd/devmigrate -file "$migration" || {
        echo "Warning: Migration $migration may have already been applied or failed"
    }
done

echo "All migrations completed"
