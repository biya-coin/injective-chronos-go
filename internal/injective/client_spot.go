package injective

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/biya-coin/injective-chronos-go/internal/model"
)

func (c *Client) SpotConfig(ctx context.Context) (*model.ChartSpotConfig, error) {
	url := fmt.Sprintf("%s%s", c.cfg.BaseURL, c.cfg.SpotConfigPath)
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
	var out model.ChartSpotConfig
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SpotMarketSummaryAll(ctx context.Context, resolution string) ([]model.SpotMarketSummary, error) {
	var out []model.SpotMarketSummary
	q := url.Values{}
	if resolution != "" {
		q.Set("resolution", resolution)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s", c.cfg.BaseURL, c.cfg.SpotSummaryAllPath)+"?"+q.Encode(), nil)
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

func (c *Client) SpotMarketSummary(ctx context.Context, market string) (*model.SpotMarketSummary, error) {
	var out model.SpotMarketSummary
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, c.cfg.SpotSummaryPath)
	q := url.Values{}
	if market != "" {
		q.Set("market", market)
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

func (c *Client) SpotMarketSummaryAtResolution(ctx context.Context, market string, resolution string) (*model.SpotMarketSummary, error) {
	var out model.SpotMarketSummary
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, c.cfg.SpotSummaryPath)
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
