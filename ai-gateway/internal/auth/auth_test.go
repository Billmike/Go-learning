package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kayodeayelegun/ai-gateway/internal/auth"
)

func TestValidateBearer(t *testing.T) {
	t.Parallel()

	const expected = "secret-token"

	tests := []struct {
		name    string
		header  string
		wantErr error
	}{
		{
			name:    "missing header",
			header:  "",
			wantErr: auth.ErrMissingToken,
		},
		{
			name:    "missing bearer prefix",
			header:  "secret-token",
			wantErr: auth.ErrInvalidToken,
		},
		{
			name:    "empty bearer token",
			header:  "Bearer ",
			wantErr: auth.ErrInvalidToken,
		},
		{
			name:    "wrong token",
			header:  "Bearer wrong-token",
			wantErr: auth.ErrInvalidToken,
		},
		{
			name:    "valid token",
			header:  "Bearer secret-token",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			err := auth.ValidateBearer(req, expected)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ValidateBearer() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateBearer() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "missing header", header: "", want: ""},
		{name: "malformed header", header: "Basic abc", want: ""},
		{name: "valid bearer", header: "Bearer my-key", want: "my-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			if got := auth.ExtractToken(req); got != tt.want {
				t.Fatalf("ExtractToken() = %q, want %q", got, tt.want)
			}
		})
	}
}
