package sentryx

import (
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
)

// Init configures the global Sentry client. Returns false when dsn is empty or init fails.
func Init(dsn, environment string) bool {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return false
	}

	opts := sentry.ClientOptions{
		Dsn:              dsn,
		TracesSampleRate: 0.1,
	}
	if env := strings.TrimSpace(environment); env != "" {
		opts.Environment = env
	}

	if err := sentry.Init(opts); err != nil {
		log.Printf("WARNING: Sentry initialization failed: %v", err)
		return false
	}

	return true
}

// Middleware returns the Fiber handler that attaches a per-request Sentry hub.
// Register after recover middleware with Repanic enabled so panics are captured
// and then handled by Fiber's recover handler.
func Middleware() fiber.Handler {
	return sentryfiber.New(sentryfiber.Options{
		Repanic:         true,
		WaitForDelivery: false,
	})
}

// Flush waits for pending Sentry events before shutdown.
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}
