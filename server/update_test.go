package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOnlineUpdateDisabled(t *testing.T) {
	var srv Server
	cases := []struct {
		name   string
		method string
		path   string
		handle func(http.ResponseWriter, *http.Request)
		want   int
	}{
		{"check", http.MethodGet, "/api/update/check", srv.updateCheck, http.StatusOK},
		{"apply", http.MethodPost, "/api/update/apply", srv.updateApply, http.StatusGone},
		{"rollback", http.MethodPost, "/api/update/rollback", srv.updateRollback, http.StatusGone},
		{"stream", http.MethodGet, "/api/update/stream", srv.updateStream, http.StatusGone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.handle(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tc.want, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), onlineUpdateDisabled) {
				t.Fatalf("body %s does not contain the disabled message", rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), "github.com") {
				t.Fatalf("body names a GitHub host: %s", rec.Body.String())
			}
		})
	}
}

func TestUpdateCheckIsLocalVersionOnly(t *testing.T) {
	prev := BuildVersion
	BuildVersion = "0.3.15"
	t.Cleanup(func() { BuildVersion = prev })

	rec := httptest.NewRecorder()
	var srv Server
	srv.updateCheck(rec, httptest.NewRequest(http.MethodGet, "/api/update/check", nil))

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["current"] != "0.3.15" {
		t.Fatalf("current = %#v", body["current"])
	}
	if body["online_update"] != false {
		t.Fatalf("online_update = %#v", body["online_update"])
	}
	if body["message"] != onlineUpdateDisabled {
		t.Fatalf("message = %#v", body["message"])
	}
	for _, key := range []string{"latest", "html_url", "repo", "asset", "has_update", "notes"} {
		if _, ok := body[key]; ok {
			t.Fatalf("check response still has %s", key)
		}
	}
}
