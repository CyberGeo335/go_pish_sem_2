package httpapi

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "health", path: "/health", want: "/health"},
		{name: "metrics", path: "/metrics", want: "/metrics"},
		{name: "student id", path: "/students/1", want: "/students/{id}"},
		{name: "students empty", path: "/students", want: "/students/{id}"},
		{name: "unknown", path: "/unknown", want: "/unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePath(tt.path)
			if got != tt.want {
				t.Fatalf("normalizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
