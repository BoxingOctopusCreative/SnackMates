package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/boxingoctopus/snackmates/api/internal/config"
	"github.com/boxingoctopus/snackmates/api/internal/db"
	"github.com/boxingoctopus/snackmates/api/internal/search"
	"github.com/boxingoctopus/snackmates/api/internal/seed"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

func main() {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "seed",
		Short: "Manage example users and wishlists in the database",
	}

	var password string
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Seed the database with example users and wishlists",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := openEnv(configFile)
			if err != nil {
				return err
			}
			defer env.pool.Close()
			return seed.Run(env.ctx, env.pool, env.search, password)
		},
	}
	runCmd.Flags().StringVar(&password, "password", seed.DefaultPassword, "password for seeded users")

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove seeded example users and wishlists",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := openEnv(configFile)
			if err != nil {
				return err
			}
			defer env.pool.Close()
			return seed.Remove(env.ctx, env.pool, env.search)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to TOML config file (falls back to CONFIG_FILE env var)")
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.RunE = runCmd.RunE

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type seedEnv struct {
	ctx    context.Context
	pool   *pgxpool.Pool
	search *search.Client
}

func openEnv(configFile string) (*seedEnv, error) {
	cfg, err := config.Load(config.Options{
		ConfigFile: config.ResolveConfigPath(configFile),
	})
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	migrationsDir := db.ResolveMigrationsDir()
	if err := db.RunMigrations(cfg.DatabaseURL, migrationsDir); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations (%s): %w", migrationsDir, err)
	}
	log.Printf("database migrations up to date (%s)", migrationsDir)

	return &seedEnv{
		ctx:    ctx,
		pool:   pool,
		search: connectMeilisearch(cfg.MeilisearchURL, cfg.MeilisearchAPIKey),
	}, nil
}

func connectMeilisearch(url, apiKey string) *search.Client {
	client, err := search.New(url, apiKey)
	if err != nil {
		log.Printf("WARNING: meilisearch unavailable; skipping wishlist indexing (%v)", err)
		return search.Disabled
	}
	return client
}
