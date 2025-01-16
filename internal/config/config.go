package config

import (
	"log/slog"
)

type DB struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type Server struct {
	Port int
}

type RemoteAPI struct {
	URL string
}

type Logger struct {
	Level     slog.Level
	PlainText bool
}

type Config struct {
	Server    Server
	DB        DB
	RemoteAPI RemoteAPI
	Logger    Logger
}

func Load() (Config, error) {

	var ge getenv
	const required = true

	cfg := Config{
		Server: Server{
			Port: ge.Int("SERVER_PORT", required, 8080),
		},
		DB: DB{
			Host:     ge.String("DB_HOST", !required, "localhost"),
			Port:     ge.Int("DB_PORT", !required, 5432),
			Name:     ge.String("DB_NAME", required, "postgres"),
			User:     ge.String("DB_USER", required, "postgres"),
			Password: ge.String("DB_PASS", required, "postgres"),
		},
		RemoteAPI: RemoteAPI{
			URL: ge.String("REMOTE_API_URL", required, "http://localhost:8081"),
		},
		Logger: Logger{
			Level:     ge.LogLevel("LOG_LEVEL", !required, slog.LevelInfo),
			PlainText: ge.Bool("LOG_PLAINTEXT", !required, false),
		},
	}

	return cfg, ge.err
}
