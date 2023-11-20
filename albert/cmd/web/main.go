package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

	"albert/assets"
	"albert/internal/configs"
	"albert/internal/home"
	"albert/internal/migrations"
	"albert/internal/securities"
	"albert/internal/users"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var settings configs.Settings
	if err := envconfig.Process(configs.AppName, &settings); err != nil {
		panic(fmt.Errorf("failed to parse environment variables: %w", err))
	}

	var pool *pgxpool.Pool
	{
		var err error
		pool, err = pgxpool.New(ctx, settings.DatabaseUrl)
		if err != nil {
			panic(fmt.Errorf("failed to connect to database: %w", err))
		}
		if err = migrations.Run(ctx, pool); err != nil {
			panic(fmt.Errorf("failed to run database migrations: %w", err))
		}
	}
	defer pool.Close()

	sessionManager := scs.New()
	sessionManager.Cookie.Name = configs.AppName
	sessionManager.Cookie.Persist = true
	sessionManager.Lifetime = settings.SessionMaxLifetime
	sessionManager.Store = pgxstore.NewWithCleanupInterval(pool, settings.SessionCleanupInterval)

	homeHandler := home.NewHandler(logger)
	securitiesHandler := securities.NewHandler(settings.ApiKey, pool, logger)
	userHandler := users.NewHandler(pool, sessionManager, logger)
	csrfMiddleware := csrf.Protect(settings.CsrfKey, csrf.FieldName(configs.CsrfFormFieldName))

	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Heartbeat("/health"))
	mux.Use(csrfMiddleware)
	mux.Use(userHandler.WithCurrentUser)
	assets.Mount(mux)
	mux.Get("/", homeHandler.ShowIndex)
	mux.Route("/auth", func(r chi.Router) {
		r.Get("/login", userHandler.ShowLogin)
		r.Post("/login", userHandler.DoLogin)
		r.Post("/logout", userHandler.DoLogout)
	})
	mux.Route("/watchlist", func(r chi.Router) {
		r.Use(users.RequireLogin)
		r.Get("/", userHandler.ShowWatchlist)
		r.Post("/add", userHandler.Watch)
		r.Post("/remove", userHandler.Unwatch)
	})
	mux.Route("/securities", func(r chi.Router) {
		r.Use(users.RequireLogin)
		r.Post("/", securitiesHandler.Search)
	})
	srv := http.Server{Addr: settings.WebAddress, Handler: sessionManager.LoadAndSave(mux)}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		logger.Info("starting server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start web server: %w", err)
		}
		return nil
	})

	group.Go(func() error {
		<-groupCtx.Done()
		newCtx, cancel := context.WithTimeout(context.Background(), settings.ShutdownWindow)
		defer cancel()
		_ = srv.Shutdown(newCtx)
		return nil
	})

	if err := group.Wait(); err != nil {
		panic(err)
	}
}
