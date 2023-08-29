DUCKDB_BIN := /path/to/duckdb
DB_NAME := data.db

.PHONY: create-db
create-db:
	@echo "Creating DuckDB database and tables..."
	$(DUCKDB_BIN) $(DB_NAME) <<EOF
	CREATE TABLE IF NOT EXISTS id_sequence (entity_name VARCHAR, next_id INTEGER);
EOF
	@echo "DuckDB database and tables created successfully."

.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f $(DB_NAME)
	@echo "Cleanup complete."
