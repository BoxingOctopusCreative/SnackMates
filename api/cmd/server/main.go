package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/cache"
	"github.com/boxingoctopus/snackmates/api/internal/config"
	"github.com/boxingoctopus/snackmates/api/internal/db"
	"github.com/boxingoctopus/snackmates/api/internal/handlers"
	"github.com/boxingoctopus/snackmates/api/internal/notifications"
	"github.com/boxingoctopus/snackmates/api/internal/search"
	"github.com/boxingoctopus/snackmates/api/internal/snacksearch"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/cobra"
)

func main() {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "server",
		Short: "Run the SnackMates API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(configFile)
		},
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "path to TOML config file (falls back to CONFIG_FILE env var)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(configFile string) error {
	cfg, err := config.Load(config.Options{
		ConfigFile: config.ResolveConfigPath(configFile),
	})
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()

	migrationsDir := db.ResolveMigrationsDir()
	if err := db.RunMigrations(cfg.DatabaseURL, migrationsDir); err != nil {
		return fmt.Errorf("migrations (%s): %w", migrationsDir, err)
	}
	log.Printf("database migrations up to date (%s)", migrationsDir)

	valkey := connectValkey(cfg.ValkeyURL)
	defer valkey.Close()

	meili := connectMeilisearch(cfg.MeilisearchURL, cfg.MeilisearchAPIKey)
	snackSearch := snacksearch.New(snacksearch.Options{
		AnthropicAPIKey: cfg.AnthropicAPIKey,
		AnthropicModel:  cfg.AnthropicModel,
		OVHAIToken:      cfg.OVHAIToken,
		OVHAIModel:      cfg.OVHAIModel,
		OVHAIBaseURL:    cfg.OVHAIBaseURL,
		OFFSearchURL:    cfg.OFFSearchURL,
		OFFBaseURL:      cfg.OFFBaseURL,
	})

	s3, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("s3: %w", err)
	}

	discord := auth.NewDiscordService(cfg)
	webauthn, err := auth.NewWebAuthnService(cfg)
	if err != nil {
		return fmt.Errorf("webauthn: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "SnackMates API",
		BodyLimit:    10 * 1024 * 1024,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.WebOrigin,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-WebAuthn-Session, X-Device-Name",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "snackmates-api"})
	})

	notificationHub := notifications.NewHub(valkey)
	notificationHub.Start(ctx)

	v1 := app.Group("/api/v1")
	authHandler := handlers.NewAuthHandler(cfg, pool, valkey, s3, discord, webauthn)
	authHandler.RegisterRoutes(v1)
	handlers.NewWishlistHandler(pool, meili, s3).RegisterRoutes(v1)
	handlers.NewMatchHandler(pool, s3).RegisterRoutes(v1)
	handlers.NewUserHandler(pool, s3).RegisterRoutes(v1)
	handlers.NewFriendHandler(pool, s3, notificationHub).RegisterRoutes(v1)
	handlers.NewMessageHandler(pool, s3, notificationHub).RegisterRoutes(v1)
	handlers.NewChatHandler(pool, s3, notificationHub).RegisterRoutes(v1)
	handlers.NewNotificationHandler(pool, s3, notificationHub).RegisterRoutes(v1)
	handlers.NewSearchHandler(pool, meili, s3, snackSearch).RegisterRoutes(v1)

	go func() {
		addr := ":" + cfg.Port
		log.Printf("SnackMates API listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	return app.ShutdownWithTimeout(10 * time.Second)
}

const startupConnectRetries = 2

func connectValkey(url string) *cache.Client {
	var lastErr error
	for attempt := 0; attempt <= startupConnectRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}
		client, err := cache.New(url)
		if err == nil {
			return client
		}
		lastErr = err
		log.Printf("valkey connection attempt %d/%d failed: %v", attempt+1, startupConnectRetries+1, err)
	}
	log.Printf("WARNING: valkey unavailable after %d attempts; continuing without cache (%v)", startupConnectRetries+1, lastErr)
	return cache.Disabled
}

func connectMeilisearch(url, apiKey string) *search.Client {
	var lastErr error
	for attempt := 0; attempt <= startupConnectRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}
		client, err := search.New(url, apiKey)
		if err == nil {
			return client
		}
		lastErr = err
		log.Printf("meilisearch connection attempt %d/%d failed: %v", attempt+1, startupConnectRetries+1, err)
	}
	log.Printf("WARNING: meilisearch unavailable after %d attempts; continuing without search indexing (%v)", startupConnectRetries+1, lastErr)
	return search.Disabled
}
