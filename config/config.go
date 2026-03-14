package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	RedisAddr   string `envconfig:"REDIS_ADDR"   default:"localhost:6379"`
	Domain      string `envconfig:"DOMAIN"       required:"true"`
	APIQuota    int    `envconfig:"API_QUOTA"    default:"10"`
	Port        string `envconfig:"PORT"         default:"8080"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
