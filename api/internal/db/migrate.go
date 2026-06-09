package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies pending SQL migrations from dir against databaseURL.
func RunMigrations(databaseURL, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve migrations dir: %w", err)
	}

	if st, err := os.Stat(absDir); err != nil {
		return fmt.Errorf("migrations dir %q: %w", absDir, err)
	} else if !st.IsDir() {
		return fmt.Errorf("migrations path %q is not a directory", absDir)
	}

	sourceURL := "file://" + filepath.ToSlash(absDir)
	m, err := migrate.New(sourceURL, toMigrateURL(databaseURL))
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		_, _ = m.Close()
		return fmt.Errorf("apply migrations: %w", err)
	}

	if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
		return errors.Join(
			fmt.Errorf("close migration source: %w", srcErr),
			fmt.Errorf("close migration database: %w", dbErr),
		)
	}
	return nil
}

// ResolveMigrationsDir picks a migrations directory from MIGRATIONS_DIR or common paths.
func ResolveMigrationsDir() string {
	if dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); dir != "" {
		return dir
	}

	if dir, ok := findMigrationsDirFrom(os.Getwd); ok {
		return dir
	}

	for _, candidate := range []string{"migrations", "api/migrations"} {
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate
		}
	}
	return "migrations"
}

func findMigrationsDirFrom(getwd func() (string, error)) (string, bool) {
	cwd, err := getwd()
	if err != nil {
		return "", false
	}

	dir := cwd
	for {
		if st, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !st.IsDir() {
			migrations := filepath.Join(dir, "migrations")
			if st, err := os.Stat(migrations); err == nil && st.IsDir() {
				return migrations, true
			}
		}

		apiMigrations := filepath.Join(dir, "api", "migrations")
		if st, err := os.Stat(apiMigrations); err == nil && st.IsDir() {
			return apiMigrations, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", false
}

func toMigrateURL(databaseURL string) string {
	switch {
	case strings.HasPrefix(databaseURL, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(databaseURL, "postgres://")
	case strings.HasPrefix(databaseURL, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(databaseURL, "postgresql://")
	default:
		return databaseURL
	}
}
