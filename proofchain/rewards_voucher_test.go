package proofchain

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newVoucherRewardsClient starts a test server and returns a rewards client aimed at it.
func newVoucherRewardsClient(t *testing.T, handler http.HandlerFunc) *RewardsClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewRewardsClient(mustHTTPClient(t, "test-key", WithBaseURL(srv.URL)))
}

func voucherRequestBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("decode body: %v", err)
	}
	return body
}

func TestRewardsClaimVoucherAndListProducts(t *testing.T) {
	client := newVoucherRewardsClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rewards/earned/r/claim":
			if r.Method != http.MethodPost || r.URL.RawQuery != "" {
				t.Errorf("voucher claim = %s ?%q", r.Method, r.URL.RawQuery)
			}
			if raw, _ := io.ReadAll(r.Body); len(raw) != 0 {
				t.Errorf("claim body = %q", raw)
			}
			_, _ = io.WriteString(w, `{"success":true,"reward_id":"r","status":"claimed","redemption_url":"https://example.com/payout","expires_at":null}`)
		case "/rewards/voucher-products":
			if r.Method != http.MethodGet {
				t.Errorf("products method = %s", r.Method)
			}
			if r.URL.RawQuery != "category=merchant_card%2Ccharity&country=GB%2CIE&currency=GBP&fee_free=false&subcategory=food_and_drink%2Centertainment" {
				t.Errorf("products query = %q", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"Z88","name":"Argos UK","category":"merchant_card","subcategory":null,"image_url":null,"currency_codes":["GBP"],"countries":["GB"],"skus":[{"min":5.0,"max":500.0,"currency_code":"GBP"}]},{"id":"C1","name":"Charity","category":"charity","description":"d","disclosure":"<p>x</p>","usage_instructions":"u","fee_free":true,"currency_codes":["GBP"],"countries":["GB"],"skus":[]}]`)
		case "/rewards/definitions":
			body := voucherRequestBody(t, r)
			if body["voucher_language"] != "fr" || body["voucher_message"] != "Merci" {
				t.Errorf("definition body = %#v", body)
			}
			ids, _ := body["voucher_product_ids"].([]interface{})
			if len(ids) != 1 || ids[0] != "ALL_FEE_FREE" {
				t.Errorf("voucher_product_ids = %#v", body["voucher_product_ids"])
			}
			_, _ = io.WriteString(w, `{"id":"d","voucher_language":"fr","voucher_message":"Merci","voucher_product_ids":["C1"]}`)
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
		}
	})
	ctx := context.Background()
	claim, err := client.ClaimVoucher(ctx, "r", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !claim.Success || claim.Status != "claimed" || claim.RedemptionURL == nil || *claim.RedemptionURL != "https://example.com/payout" || claim.ExpiresAt != nil {
		t.Fatalf("voucher claim shape = %#v", claim)
	}
	feeFree := false
	products, err := client.ListVoucherProducts(ctx, &ListVoucherProductsParams{
		Country: "GB,IE", Currency: "GBP", Category: "merchant_card,charity",
		Subcategory: "food_and_drink,entertainment", FeeFree: &feeFree,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 || products[0].ID != "Z88" || products[0].Subcategory != nil || len(products[0].SKUs) != 1 || products[0].SKUs[0].Max != 500 {
		t.Fatalf("products shape = %#v", products)
	}
	if products[0].FeeFree || products[0].Description != nil || products[0].Disclosure != nil || products[0].UsageInstructions != nil {
		t.Fatalf("v1 product defaults = %#v", products[0])
	}
	if p := products[1]; !p.FeeFree || *p.Description != "d" || *p.Disclosure != "<p>x</p>" || *p.UsageInstructions != "u" {
		t.Fatalf("v2 product fields = %#v", p)
	}
	if VoucherAllFeeFree != "ALL_FEE_FREE" {
		t.Fatalf("VoucherAllFeeFree = %q", VoucherAllFeeFree)
	}
	lang, msg := "fr", "Merci"
	def, err := client.CreateDefinition(ctx, &CreateRewardDefinitionRequest{
		Name: "v", RewardType: "voucher", VoucherProductIDs: []string{VoucherAllFeeFree},
		VoucherLanguage: &lang, VoucherMessage: &msg,
	})
	if err != nil {
		t.Fatal(err)
	}
	if *def.VoucherLanguage != "fr" || *def.VoucherMessage != "Merci" || len(def.VoucherProductIDs) != 1 {
		t.Fatalf("definition = %#v", def)
	}
}

func TestRewardsEarnedVoucherDeliveryFields(t *testing.T) {
	want := map[string]string{
		"e": `{"delivery":{"method":"email","email":"fan@example.com","name":"Sam"}}`,
		"s": `{"delivery":{"method":"sms","phone":"+447911123456"}}`,
		"l": `{"delivery":{"method":"link"}}`,
	}
	client := newVoucherRewardsClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rewards/earned/e/claim", "/rewards/earned/s/claim", "/rewards/earned/l/claim":
			id := strings.Split(r.URL.Path, "/")[3]
			if r.Method != http.MethodPost {
				t.Errorf("claim %s method = %s", id, r.Method)
			}
			raw, _ := io.ReadAll(r.Body)
			if string(raw) != want[id] {
				t.Errorf("claim %s body = %s", id, raw)
			}
			_, _ = io.WriteString(w, `{"success":true,"reward_id":"`+id+`","status":"claimed","redemption_url":null,"expires_at":null,"delivery_method":"email","delivery_status":"pending"}`)
		case "/rewards/definitions":
			body := voucherRequestBody(t, r)
			methods, _ := body["voucher_delivery_methods"].([]interface{})
			if len(methods) != 2 || methods[0] != "link" || methods[1] != "email" {
				t.Errorf("voucher_delivery_methods = %#v", body["voucher_delivery_methods"])
			}
			_, _ = io.WriteString(w, `{"id":"d","voucher_delivery_methods":["link","email"]}`)
		case "/rewards/earned":
			_, _ = io.WriteString(w, `[{"id":"x","voucher_delivery_method":"email","voucher_delivery_status":"failed","voucher_delivery_failure":"invalid_email","voucher_status_updated_at":"2026-11-03T09:16:43.119933"},{"id":"y","voucher_delivery_method":null,"voucher_delivery_status":null,"voucher_delivery_failure":null,"voucher_status_updated_at":null}]`)
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
		}
	})
	ctx := context.Background()
	for id, d := range map[string]*VoucherDelivery{
		"e": {Method: VoucherDeliveryEmail, Email: "fan@example.com", Name: "Sam"},
		"s": {Method: VoucherDeliverySMS, Phone: "+447911123456"},
		"l": {Method: VoucherDeliveryLink},
	} {
		claim, err := client.ClaimVoucher(ctx, id, d)
		if err != nil {
			t.Fatal(err)
		}
		if claim.RedemptionURL != nil || claim.DeliveryMethod != "email" || claim.DeliveryStatus != "pending" {
			t.Fatalf("claim %s = %#v", id, claim)
		}
	}
	def, err := client.CreateDefinition(ctx, &CreateRewardDefinitionRequest{
		Name: "v", RewardType: "voucher", VoucherDeliveryMethods: []string{VoucherDeliveryLink, VoucherDeliveryEmail},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(def.VoucherDeliveryMethods) != 2 || def.VoucherDeliveryMethods[1] != "email" {
		t.Fatalf("definition = %#v", def)
	}
	earned, err := client.ListEarned(ctx, "", "", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	e := earned[0]
	if e.VoucherDeliveryMethod == nil || *e.VoucherDeliveryMethod != "email" || *e.VoucherDeliveryStatus != "failed" ||
		*e.VoucherDeliveryFailure != "invalid_email" || e.VoucherStatusUpdatedAt == nil {
		t.Fatalf("earned = %#v", e)
	}
	if want := time.Date(2026, 11, 3, 9, 16, 43, 119933000, time.UTC); !e.VoucherStatusUpdatedAt.Time.Equal(want) {
		t.Fatalf("voucher_status_updated_at = %v, want %v", e.VoucherStatusUpdatedAt.Time, want)
	}
	if n := earned[1]; n.VoucherDeliveryMethod != nil || n.VoucherDeliveryStatus != nil || n.VoucherDeliveryFailure != nil || n.VoucherStatusUpdatedAt != nil {
		t.Fatalf("earned nulls = %#v", n)
	}
}
