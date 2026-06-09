package storage

import "testing"

func TestBuildObjectURLPathStyle(t *testing.T) {
	got := buildObjectURL("https://abc123.r2.cloudflarestorage.com", "client-assets", "avatars/u1/pic.png", true)
	want := "https://abc123.r2.cloudflarestorage.com/client-assets/avatars/u1/pic.png"
	if got != want {
		t.Fatalf("buildObjectURL() = %q, want %q", got, want)
	}
}

func TestBuildObjectURLVirtualHosted(t *testing.T) {
	got := buildObjectURL("https://abc123.r2.cloudflarestorage.com", "client-assets", "avatars/u1/pic.png", false)
	want := "https://client-assets.abc123.r2.cloudflarestorage.com/avatars/u1/pic.png"
	if got != want {
		t.Fatalf("buildObjectURL() = %q, want %q", got, want)
	}
}

func TestBuildObjectURLPublicBase(t *testing.T) {
	got := buildObjectURL("https://cdn.example.com", "client-assets", "logo.png", true)
	want := "https://cdn.example.com/client-assets/logo.png"
	if got != want {
		t.Fatalf("buildObjectURL() = %q, want %q", got, want)
	}
}

func TestBuildObjectURLMinIO(t *testing.T) {
	got := buildObjectURL("http://localhost:9000", "client-assets", "avatars/test.jpg", true)
	want := "http://localhost:9000/client-assets/avatars/test.jpg"
	if got != want {
		t.Fatalf("buildObjectURL() = %q, want %q", got, want)
	}
}
