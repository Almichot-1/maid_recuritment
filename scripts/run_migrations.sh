#!/bin/bash

echo "Running database migrations..."

MIGRATION_FAILED=0

# Get all .up.sql files in order
for migration in $(ls migrations/*.up.sql 2>/dev/null | sort -V); do
    echo "Applying migration: $migration"
    if go run ./cmd/devmigrate -file "$migration"; then
        echo "OK: $migration applied"
    else
        echo "SKIP: $migration already applied or failed (continuing)"
    fi
done

echo "Migration step completed"
