package main

import "testing"

func TestTaskIDFromPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    int
		wantErr bool
	}{
		{name: "valid", path: "/tasks/12", want: 12},
		{name: "empty", path: "/tasks/", wantErr: true},
		{name: "not numeric", path: "/tasks/abc", wantErr: true},
		{name: "nested", path: "/tasks/1/comments", wantErr: true},
		{name: "zero", path: "/tasks/0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := taskIDFromPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}
