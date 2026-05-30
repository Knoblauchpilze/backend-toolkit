#!/bin/bash

echo "Setting up test database..."
echo "You may need to enter multiple times the postgres password"

export ADMIN_PASSWORD="admin_password"
export MANAGER_PASSWORD="manager_password"
export USER_PASSWORD="test_password"

echo "Creating users..."
/bin/bash database/create_user.sh database/test

echo "Creating database..."
/bin/bash database/create_database.sh database/test

echo "Seeding values..."
export DB_PATH="database/test"
make -f database/Makefile migrate

echo "All done!"
