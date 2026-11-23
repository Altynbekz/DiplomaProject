package main

import (
	"context"
	"diploma/internal/classify"
	"diploma/internal/config"
	"diploma/internal/httpapi"
	"diploma/internal/service"
	"diploma/internal/store"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func run() error {
	c, e := config.Load()
	if e != nil {
		return e
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pc, e := pgxpool.ParseConfig(c.DSN)
	if e != nil {
		return errors.New("invalid DATABASE_URL")
	}
	pc.MaxConns = 5
	pc.ConnConfig.ConnectTimeout = 3 * time.Second
	pool, e := pgxpool.NewWithConfig(ctx, pc)
	if e != nil {
		return errors.New("database configuration failed")
	}
	defer pool.Close()
	rules, e := classify.Load()
	if e != nil {
		return e
	}
	svc := &service.Service{Mapping: rules, Store: store.New(pool), Origin: c.Origin, Actor: c.Actor, CursorKey: c.Token, Now: time.Now}
	a := &httpapi.API{Service: svc, Token: c.Token, Log: slog.Default()}
	server := &http.Server{Addr: c.Listen, Handler: a.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 45 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	slog.Info("API started", "scanning_enabled", false)
	select {
	case e = <-done:
		if !errors.Is(e, http.ErrServerClosed) {
			return errors.New("HTTP server failed")
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
	return nil
}
func main() {
	if len(os.Args) == 2 && os.Args[1] == "-healthcheck" {
		client := http.Client{Timeout: 2 * time.Second}
		r, e := client.Get("http://127.0.0.1:8080/readyz")
		if e != nil {
			os.Exit(1)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			os.Exit(1)
		}
		return
	}

	if e := run(); e != nil {
		slog.Error("API stopped", "reason", e.Error())
		os.Exit(1)
	}
}
