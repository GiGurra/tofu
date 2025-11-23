package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunHttp(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, _ := io.ReadAll(r.Body)
			w.Write([]byte("received: " + string(body)))
			return
		}
		if r.Header.Get("X-Custom-Header") == "test-value" {
			w.Write([]byte("header-ok"))
			return
		}
		w.Write([]byte("hello world"))
	}))
	defer server.Close()

	tests := []struct {
		name    string
		params  HttpParams
		want    string
		wantErr bool
	}{
		{
			name: "Basic GET",
			params: HttpParams{
				URL: server.URL,
			},
			want: "hello world",
		},
		{
			name: "POST with data",
			params: HttpParams{
				URL:  server.URL,
				Data: "some data",
			},
			want: "received: some data",
		},
		{
			name: "GET with Header",
			params: HttpParams{
				URL:     server.URL,
				Headers: []string{"X-Custom-Header: test-value"},
			},
			want: "header-ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runHttp(&tt.params, &stdout, &stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("runHttp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := stdout.String(); got != tt.want {
				t.Errorf("runHttp() stdout = %v, want %v", got, tt.want)
			}
		})
	}
}
