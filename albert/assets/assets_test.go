package assets_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"albert/assets"
)

func TestMount(t *testing.T) {
	mux := chi.NewMux()
	assets.Mount(mux)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	appCss, err := os.ReadFile(filepath.Join(cwd, "./public/styles/app.css"))
	require.NoError(t, err)
	favicon, err := os.ReadFile(filepath.Join(cwd, "./public/icons/favicon.ico"))
	require.NoError(t, err)

	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantBody   string
	}{
		{
			name:       "unknown",
			req:        httptest.NewRequest(http.MethodGet, "/foobar", nil),
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
		{
			name:       "app.css without prefix",
			req:        httptest.NewRequest(http.MethodGet, "/app.css", nil),
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
		{
			name:       "app.css with wrong prefix",
			req:        httptest.NewRequest(http.MethodGet, "/assets/app.css", nil),
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
		{
			name:       "app.css with correct prefix",
			req:        httptest.NewRequest(http.MethodGet, "/assets/public/styles/app.css", nil),
			wantStatus: http.StatusOK,
			wantBody:   string(bytes.Join([][]byte{appCss}, []byte("\n"))),
		},
		{
			name:       "favicon.ico with incorrect prefix",
			req:        httptest.NewRequest(http.MethodGet, "/assets/favicon.ico", nil),
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
		{
			name:       "favicon.ico with correct prefix",
			req:        httptest.NewRequest(http.MethodGet, "/assets/public/icons/favicon.ico", nil),
			wantStatus: http.StatusOK,
			wantBody:   string(bytes.Join([][]byte{favicon}, []byte("\n"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, tt.req)

			require.Equal(t, tt.wantStatus, w.Code)
			require.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
