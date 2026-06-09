package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindMigrationsDirFrom(t *testing.T) {
	apiDir := filepath.Join("..", "..")
	absAPI, err := filepath.Abs(apiDir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		cwd  string
		want string
	}{
		{
			name: "api root",
			cwd:  absAPI,
			want: filepath.Join(absAPI, "migrations"),
		},
		{
			name: "cmd seed",
			cwd:  filepath.Join(absAPI, "cmd", "seed"),
			want: filepath.Join(absAPI, "migrations"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := findMigrationsDirFrom(func() (string, error) {
				return tt.cwd, nil
			})
			if !ok {
				t.Fatal("expected migrations dir to be found")
			}
			if got != tt.want {
				t.Fatalf("findMigrationsDirFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindMigrationsDirFromRepoRoot(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docker-compose.yml")); err != nil {
		t.Skip("repo root not found from test cwd")
	}

	got, ok := findMigrationsDirFrom(func() (string, error) {
		return repoRoot, nil
	})
	if !ok {
		t.Fatal("expected migrations dir to be found from repo root")
	}
	want := filepath.Join(repoRoot, "api", "migrations")
	if got != want {
		t.Fatalf("findMigrationsDirFrom() = %q, want %q", got, want)
	}
}

func TestToMigrateURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"postgres://user:pass@localhost/db", "pgx5://user:pass@localhost/db"},
		{"postgresql://user:pass@localhost/db", "pgx5://user:pass@localhost/db"},
		{"pgx5://already", "pgx5://already"},
	}
	for _, tt := range tests {
		if got := toMigrateURL(tt.in); got != tt.want {
			t.Fatalf("toMigrateURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
