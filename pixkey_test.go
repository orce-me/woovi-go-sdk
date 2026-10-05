package woovi_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestPixKeysCreateListDeleteCheck(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/pix-keys", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["pixKey"] != "campanha@woovi.com" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"pixKey":{"key":"campanha@woovi.com","type":"EMAIL"}}`))
	})
	mux.HandleFunc("GET /api/v1/pix-keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixKeys":[{"key":"campanha@woovi.com","type":"EMAIL"}],"account":{"accountId":"acc-1","balance":{"total":100000,"blocked":0,"available":100000}}}`))
	})
	mux.HandleFunc("DELETE /api/v1/pix-keys/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/v1/pix-keys/check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixKeyEndToEndId":"E123","pixKey":"00000000191","type":"CPF","owner":{"name":"Fulano","taxID":"000.***.***-91","psp":"12345678","branch":"****","account":"********"}}`))
	})
	mux.HandleFunc("POST /api/v1/pix-keys/withdraw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"account":{"accountId":"acc-1","pixKeyWithdraw":{"pixKey":"saque@woovi.com","type":"EMAIL"}}}`))
	})
	mux.HandleFunc("PUT /api/v1/pix-keys/withdraw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"account":{"accountId":"acc-1","pixKeyWithdraw":{"pixKey":"saque@woovi.com","type":"EMAIL"}}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)

	created, err := client.PixKeys.Create(context.Background(), &woovi.PixKeyCreateParams{
		PixKey: "campanha@woovi.com",
		Type:   woovi.PixKeyTypeEmail,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Key != "campanha@woovi.com" {
		t.Fatalf("created=%+v", created)
	}

	list, err := client.PixKeys.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.PixKeys) != 1 || list.Account == nil || list.Account.Balance.Available != 100000 {
		t.Fatalf("list=%+v", list)
	}

	checked, err := client.PixKeys.Check(context.Background(), "00000000191")
	if err != nil {
		t.Fatal(err)
	}
	if checked.PixKeyEndToEndID != "E123" || checked.Owner.PSP != "12345678" {
		t.Fatalf("check=%+v", checked)
	}

	acc, err := client.PixKeys.SetWithdrawKey(context.Background(), &woovi.PixKeyWithdrawParams{
		PixKey: "saque@woovi.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if acc.PixKeyWithdraw == nil {
		t.Fatal("expected pixKeyWithdraw")
	}

	if _, err := client.PixKeys.UpdateWithdrawKey(context.Background(), &woovi.PixKeyWithdrawParams{
		PixKey: "saque@woovi.com",
	}); err != nil {
		t.Fatal(err)
	}

	if err := client.PixKeys.Delete(context.Background(), "campanha@woovi.com"); err != nil {
		t.Fatal(err)
	}
}

func TestPixKeyCheckNotFoundCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"chave nao encontrada","errorCode":"PIX_KEY_INFO_NOT_FOUND"}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	_, err := client.PixKeys.Check(context.Background(), "00000000191")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, woovi.ErrNotFound) {
		t.Fatalf("errors.Is NotFound: %v", err)
	}
	var apiErr *woovi.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "PIX_KEY_INFO_NOT_FOUND" {
		t.Fatalf("apiErr=%+v", apiErr)
	}
}
