package woovi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecodeEMV(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/decode/emv" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["emv"] == "" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"emv":{
				"payloadFormatIndicator":"01",
				"merchantAccountInformationPix":{"gui":"br.gov.bcb.pix","pixKey":"key-1"},
				"transactionAmount":"10.00",
				"merchantName":"Fulano",
				"crc":"0C98"
			},
			"cobLocation":null,
			"recLocation":null
		}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	result, err := client.Decode.EMV(context.Background(), "000201010212...")
	if err != nil {
		t.Fatal(err)
	}
	if result.EMV.MerchantName != "Fulano" {
		t.Fatalf("name=%q", result.EMV.MerchantName)
	}
	if result.EMV.MerchantAccountInformationPix == nil || result.EMV.MerchantAccountInformationPix.PixKey != "key-1" {
		t.Fatalf("pix=%+v", result.EMV.MerchantAccountInformationPix)
	}
}
