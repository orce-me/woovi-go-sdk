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

func TestPartnersListGetCreate(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/partner/company", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"company_list":[{"company_id":"cmp_1","company_name":"Affiliate"}]}`))
	})
	mux.HandleFunc("GET /api/v1/partner/company/{taxID}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"company_details":{"company_id":"cmp_1","company_name":"Affiliate","tax_id":"12345678000199"}}`))
	})
	mux.HandleFunc("POST /api/v1/partner/application", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["preRegistration"] == nil {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	mux.HandleFunc("GET /api/v1/partner/affiliate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"affiliates":[{"company":{"id":"c1","name":"Affiliate Co","taxID":"65914571000187"},"account":{"accountId":"acc1","branch":"0001","account":"1234567"}}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	list, err := client.Partners.ListCompanies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].CompanyID != "cmp_1" {
		t.Fatalf("list=%+v", list)
	}

	got, err := client.Partners.GetCompany(context.Background(), "12345678000199")
	if err != nil {
		t.Fatal(err)
	}
	if got.TaxID != "12345678000199" {
		t.Fatalf("got=%+v", got)
	}

	created, err := client.Partners.CreateApplication(context.Background(), &woovi.PartnerCreateParams{
		PreRegistration: woovi.PartnerPreRegistration{
			Name:    "Example LLC",
			TaxID:   woovi.PartnerTaxID{TaxID: "11111111111111", Type: "BR:CNPJ"},
			Website: "examplellc.com",
		},
		User: woovi.PartnerUser{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@examplellc.com",
			Phone:     "+5511912345678",
			TaxID:     woovi.PartnerTaxID{TaxID: "1111111111", Type: "BR:CPF"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.PreRegistration.Name != "Example LLC" {
		t.Fatalf("created=%+v", created)
	}

	affiliates, err := client.Partners.ListAffiliates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(affiliates.Affiliates) != 1 || affiliates.Affiliates[0].Company.Name != "Affiliate Co" {
		t.Fatalf("affiliates=%+v", affiliates)
	}
}
