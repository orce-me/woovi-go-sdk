package woovi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestApplicationsCreate(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/application" || r.Method != http.MethodPost {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["accountId"] != "acc-1" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"application":{"name":"My Test Application","isActive":true,"type":"API","clientId":"client_123","clientSecret":"secret_456","appID":"app-xyz","companyBankAccount":"acc-1"}}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	app, err := client.Applications.Create(context.Background(), &woovi.ApplicationCreateParams{
		AccountID: "acc-1",
		Application: woovi.ApplicationInput{
			Name: "My Test Application",
			Type: woovi.ApplicationTypeAPI,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if app.AppID != "app-xyz" {
		t.Fatalf("appID=%q", app.AppID)
	}
	if app.ClientSecret == "" {
		t.Fatal("expected clientSecret")
	}
}

func TestApplicationAppIdLowercase(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"application":{"name":"x","isActive":true,"type":"API","appId":"lower-id"}}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	app, err := client.Applications.Create(context.Background(), &woovi.ApplicationCreateParams{
		AccountID:   "acc-1",
		Application: woovi.ApplicationInput{Name: "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if app.AppID != "lower-id" {
		t.Fatalf("appID=%q", app.AppID)
	}
}
