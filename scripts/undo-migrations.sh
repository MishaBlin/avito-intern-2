#!/bin/bash

set -e

echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q'; do
  >&2 echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

>&2 echo "PostgreSQL is up - undoing migrations"

MIGRATIONS_DIR="/app/migrations"
MIGRATION_FILES=$(ls -1 $MIGRATIONS_DIR/*.down.sql | sort)

for migration in $MIGRATION_FILES; do
  echo "Undoing migration: $migration"
  PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$migration"
done

echo "All migrations undone successfully"