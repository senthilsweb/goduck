package config

//QueueConfiguration  Configuration
type QueueConfiguration struct {
	Broker            string `default:"redis"`
	Uri               string `default:""`
	Host              string `default:"127.0.0.1"`
	Port              string `default:"6379"`
	Password          string `default:""`
	ResultBackendHost string `default:"127.0.0.1"`
	ResultBackendPort string `default:"6379"`
}
