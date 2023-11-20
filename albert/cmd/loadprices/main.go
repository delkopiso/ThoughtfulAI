package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

	"albert/internal/configs"
	"albert/internal/securities"
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
	}
	defer pool.Close()

	securitiesHandler := securities.NewHandler(settings.ApiKey, pool, logger)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		logger.Info("starting prices loader", "frequency", settings.LoadPricesFrequency, "maxAge", settings.PriceMaxAge)
		securitiesHandler.LoadPrices(groupCtx, settings.LoadPricesFrequency, settings.PriceMaxAge)
		return nil
	})

	if err := group.Wait(); err != nil {
		panic(err)
	}
}
