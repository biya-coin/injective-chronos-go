package injective

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/model"
)

func (c *Client) DerivativeMarketSummaryAll(ctx context.Context, resolution string) ([]model.DerivativeMarketSummary, error) {
	var out []model.DerivativeMarketSummary
	q := url.Values{}
	if resolution != "" {
		q.Set("resolution", resolution)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.DerivativeSummaryAllPath)+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DerivativeMarketSummaryAtResolution(ctx context.Context, market string, resolution string) (*model.DerivativeMarketSummary, error) {
	var out model.DerivativeMarketSummary
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.DerivativeSummaryPath)
	q := url.Values{}
	q.Set("indexPrice", "false")
	if market != "" {
		q.Set("marketId", market)
	}
	if resolution != "" {
		q.Set("resolution", resolution)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DerivativeConfig(ctx context.Context) (*model.ChartDerivativeConfig, error) {
	url := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.DerivativeConfigPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out model.ChartDerivativeConfig
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DerivativeSymbolInfo(ctx context.Context, group string) (*model.DerivativeSymbolInfo, error) {
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.DerivativeSymbolInfoPath)
	q := url.Values{}
	if group != "" {
		q.Set("group", group)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out model.DerivativeSymbolInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DerivativeSymbols(ctx context.Context, symbol string) (*model.DerivativeSymbolsRaw, error) {
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.DerivativeSymbolsPath)
	q := url.Values{}
	q.Set("symbol", symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out model.DerivativeSymbolsRaw
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
