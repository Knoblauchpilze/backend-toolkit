#!/bin/bash

DB_PATH=$1
DB_HOST=${DATABASE_HOST:-localhost}
DB_PORT=${DATABASE_PORT:-5432}
DB_USER=${DATABASE_USER:-postgres}

if [ "${DB_PATH}" == "" ]; then
  echo "No path provided, defaulting to test"
  DB_PATH="test"
fi

echo "Dropping database ${DB_PATH}..."
bash drop_database.sh test
echo "Creating database ${DB_PATH}..."
bash create_database.sh test
echo "Migrating database ${DB_PATH}..."
make migrate
