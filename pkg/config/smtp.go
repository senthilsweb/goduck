package config

//SMTPConfiguration Config
type SMTPConfiguration struct {
	Host        string `default:"smtp.gmail.com"`
	Port        string `default:"587"`
	Username    string
	Password    string
	SenderEmail string
	SenderName  string
	APIKey      string
	APIToken    string
	APISecret   string
}
