package config

// QueueConfiguration  Configuration
type DuckdbConfiguration struct {
	Seed_data_base_url    string   `default:""`
	Seed_tables_from_file string   `default:""`
	Ddl_queries           []string `default:""`
	Conn                  string   `default:""`
}
