package config

type SecretsStore struct {
	Enable    bool
	Engine    string `default:"vault"`
	TokenPath string `default:"/etc/systemd/system/{{app}}/override.conf"`
	IPAndPort string `default:"http://127.0.0.1:7777"`
}
