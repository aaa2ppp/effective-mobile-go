package config

import (
	"log/slog"
)

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type ServerConfig struct {
	Port int
}

type RemoteAPIConfig struct {
	URL string
}

type LoggerConfig struct {
	Level     slog.Level
	PlainText bool
}

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	RemoteAPI RemoteAPIConfig
	Logger    LoggerConfig
}

func Load() (Config, error) {

	var ge getenv
	const required = true

	cfg := Config{
		Server: ServerConfig{
			Port: ge.Int("SERVER_PORT", required, 8080),
		},
		DB: DBConfig{
			Host:     ge.String("DB_HOST", !required, "localhost"),
			Port:     ge.Int("DB_PORT", !required, 5432),
			Name:     ge.String("DB_NAME", required, "postgres"),
			User:     ge.String("DB_USER", required, "postgres"),
			Password: ge.String("DB_PASS", required, "postgres"),
		},
		RemoteAPI: RemoteAPIConfig{
			URL: ge.String("REMOTE_API_URL", required, "http://localhost:8081"),
		},
		Logger: LoggerConfig{
			Level:     ge.LogLevel("LOG_LEVEL", !required, slog.LevelInfo),
			PlainText: ge.Bool("LOG_PLAINTEXT", !required, false),
		},
	}

	return cfg, ge.err
}
