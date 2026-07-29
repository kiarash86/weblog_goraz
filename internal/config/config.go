package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	DatabaseURL string `env:"DATABASE_URL" env-default:"postgres://postgres:postgres@localhost:5432/weblog?sslmode=disable"`
	JWTKey   string `env:"JWT_KEY" env-default:"kiakiakiakia"`
	Port        string `env:"PORT" env-default:"8080"`
}

func Load() (*Config, error) {
	var config Config
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil

}
