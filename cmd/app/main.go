package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"effective-mobile-go/docs"
	"effective-mobile-go/internal/config"
	"effective-mobile-go/internal/handler"
	"effective-mobile-go/internal/middleware"
	"effective-mobile-go/internal/repo/fake_remoterepo"
	"effective-mobile-go/internal/repo/localrepo"
	"effective-mobile-go/internal/repo/remoterepo"
	"effective-mobile-go/internal/service"
)

// main godoc
//
//	@title			Song Library
//	@version		1.0
//	@license.name	Apache 2.0
//	@BasePath		/api/v1
func main() {

	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logFatal("can't load config", err)
	}

	setupLogger(cfg.Logger)

	db, err := openDB(cfg.DB)
	if err != nil {
		logFatal("can't open db", err)
	}
	defer db.Close()

	// up migrations
	if err := goose.Up(db, "migrations"); err != nil {
		logFatal("can't up migrations", err)
	}

	localRepo := localrepo.New(db)

	var remoteRepo service.RemoteRepo
	if _, ok := os.LookupEnv("FAKEREMOTE"); ok {
		remoteRepo = fake_remoterepo.New(cfg.RemoteAPI)
	} else {
		remoteRepo = remoterepo.New(cfg.RemoteAPI)
	}

	service := service.New(localRepo, remoteRepo)

	// setup router
	router := http.NewServeMux()
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.Server.Port)
	router.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://"+docs.SwaggerInfo.Host+"/swagger/doc.json")))
	router.Handle("/api/v1/", http.StripPrefix("/api/v1", handler.New(service)))

	server := setupHTTPServer(middleware.Logging(router), cfg.Server)

	// setup graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		sig := <-c
		slog.Info("signal was received", slog.Any("signal", sig))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // TODO: timeout to config
		defer cancel()
		server.Shutdown(ctx)
	}()

	slog.Info("server startup", "addr", server.Addr)
	slog.Debug("server startup", "server", fmt.Sprintf("%+v", server))

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logFatal("server failed", err)
	}

	slog.Info("server stopped")
}

func logFatal(msg string, err error) {
	slog.Log(context.Background(), slog.LevelError+2, msg, slog.Any("error", err))
	os.Exit(1)
}

func setupLogger(cfg config.Logger) {

	var h slog.Handler
	if cfg.PlainText {
		h = slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				Level: cfg.Level,
			},
		)
	} else {
		h = slog.NewJSONHandler(
			os.Stderr,
			&slog.HandlerOptions{
				Level: cfg.Level,
			},
		)
	}

	slog.SetDefault(slog.New(h))
}

func openDB(cfg config.DB) (*sql.DB, error) {
	const op = "openDB"

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}

func setupHTTPServer(handler http.Handler, cfg config.Server) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ErrorLog:     nil,              // we use the default logger
		IdleTimeout:  time.Minute,      // TODO: to config
		ReadTimeout:  10 * time.Second, // TODO: to config
		WriteTimeout: 30 * time.Second, // TODO: to config
	}
}
