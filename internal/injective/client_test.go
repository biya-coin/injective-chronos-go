package injective

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biya-coin/injective-chronos-go/internal/config"
	"github.com/biya-coin/injective-chronos-go/internal/model"
)

func newTestServer(t *testing.T, expectedMethod, expectedPath string, wantBody any, respond any, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Fatalf("method = %s, want %s", r.Method, expectedMethod)
		}
		if r.URL.Path != expectedPath {
			t.Fatalf("path = %s, want %s", r.URL.Path, expectedPath)
		}
		if wantBody != nil {
			body, _ := io.ReadAll(r.Body)
			defer r.Body.Close()
			var got any
			_ = json.Unmarshal(body, &got)
			var want any
			b, _ := json.Marshal(wantBody)
			_ = json.Unmarshal(b, &want)
			if string(b) != string(body) {
				t.Fatalf("body = %s, want %s", string(body), string(b))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(respond)
	}))
}

func TestClient_SpotMarketSummaryAll(t *testing.T) {
	resp := []model.MarketSummaryCommon{{MarketID: "m1"}}
	ts := newTestServer(t, http.MethodGet, "/spot-all", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:            ts.URL,
		SpotSummaryAllPath: "/spot-all",
	}, ts.Client())

	out, err := c.SpotMarketSummaryAll(context.Background(), "24h")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(out) == 0 || out[0].MarketID != "m1" {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestClient_SpotMarketSummary(t *testing.T) {
	resp := model.MarketSummaryCommon{MarketID: "ETH-USDT"}
	ts := newTestServer(t, http.MethodGet, "/spot", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:         ts.URL,
		SpotSummaryPath: "/spot",
	}, ts.Client())

	out, err := c.SpotMarketSummary(context.Background(), "ETH-USDT")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out == nil || out.MarketID != "ETH-USDT" {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestClient_SpotMarketSummaryAtResolution(t *testing.T) {
	resp := model.MarketSummaryCommon{MarketID: "ETH-USDT"}
	ts := newTestServer(t, http.MethodGet, "/spot", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:         ts.URL,
		SpotSummaryPath: "/spot",
	}, ts.Client())

	out, err := c.SpotMarketSummaryAtResolution(context.Background(), "ETH-USDT", "24h")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out == nil || out.MarketID != "ETH-USDT" {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestClient_DerivativeMarketSummaryAll(t *testing.T) {
	resp := []model.MarketSummaryCommon{{MarketID: "d1"}}
	ts := newTestServer(t, http.MethodGet, "/deriv-all", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:                  ts.URL,
		DerivativeSummaryAllPath: "/deriv-all",
	}, ts.Client())

	out, err := c.DerivativeMarketSummaryAll(context.Background(), "24h")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(out) == 0 || out[0].MarketID != "d1" {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestClient_DerivativeMarketSummaryAtResolution(t *testing.T) {
	resp := model.MarketSummaryCommon{MarketID: "PERP-ETH"}
	ts := newTestServer(t, http.MethodGet, "/deriv", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:               ts.URL,
		DerivativeSummaryPath: "/deriv",
	}, ts.Client())

	out, err := c.DerivativeMarketSummaryAtResolution(context.Background(), "PERP-ETH", "24h")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out == nil || out.MarketID != "PERP-ETH" {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestClient_SpotConfig(t *testing.T) {
	resp := map[string]any{
		"supported_resolutions":    []string{"1", "5", "24h"},
		"supports_group_request":   false,
		"supports_marks":           false,
		"supports_search":          true,
		"supports_timescale_marks": false,
	}
	ts := newTestServer(t, http.MethodGet, "/spot-config", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:        ts.URL,
		SpotConfigPath: "/spot-config",
	}, ts.Client())

	out, err := c.SpotConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out == nil || len(out.SupportedResolutions) == 0 {
		t.Fatalf("missing supported_resolutions: %#v", out)
	}
}

func TestClient_DerivativeConfig(t *testing.T) {
	resp := map[string]any{
		"supported_resolutions":    []string{"1", "5", "24h"},
		"supports_group_request":   false,
		"supports_marks":           false,
		"supports_search":          true,
		"supports_timescale_marks": false,
	}
	ts := newTestServer(t, http.MethodGet, "/deriv-config", nil, resp, http.StatusOK)
	defer ts.Close()

	c := NewClient(config.InjectiveConf{
		BaseURL:              ts.URL,
		DerivativeConfigPath: "/deriv-config",
	}, ts.Client())

	out, err := c.DerivativeConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out == nil || len(out.SupportedResolutions) == 0 {
		t.Fatalf("missing supported_resolutions: %#v", out)
	}
}
