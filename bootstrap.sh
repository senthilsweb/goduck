#!/bin/bash

# Specify the path to the DuckDB binary
DUCKDB_BIN="/path/to/duckdb"

# Database name
DB_NAME="data.db"

# SQL statements for table creation and sequence
SQL_CREATE_TABLE="CREATE TABLE IF NOT EXISTS id_sequence (entity_name VARCHAR, next_id INTEGER);"

# Run the DuckDB binary to create the database and tables
$DUCKDB_BIN "$DB_NAME" <<EOF
$SQL_CREATE_TABLE
EOF

# Check if the DuckDB binary exited successfully
if [ $? -eq 0 ]; then
  echo "DuckDB database and tables created successfully."
else
  echo "Error creating DuckDB database and tables."
fi
