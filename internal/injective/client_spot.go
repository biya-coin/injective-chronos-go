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
	"github.com/zeromicro/go-zero/core/logx"
)

func (c *Client) SpotConfig(ctx context.Context) (*model.ChartSpotConfig, error) {
	url := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotConfigPath)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotSummaryAllPath)+"?"+q.Encode(), nil)
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
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotSummaryPath)
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
	u := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotSummaryPath)
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

// SpotMarketHistory fetches spot candle history; same request format as MarketHistory.
// It reuses the generic MarketHistory endpoint to avoid duplication.
func (c *Client) SpotMarketHistory(ctx context.Context, from int64, to int64, marketId string, resolution string, countback int) (model.SpotMarketHistory, error) {
	if marketId == "" {
		return model.SpotMarketHistory{}, fmt.Errorf("SpotMarketHistory request marketIDs is required")
	}
	if resolution == "" {
		return model.SpotMarketHistory{}, fmt.Errorf("SpotMarketHistory request resolution is required")
	}

	q := url.Values{}
	if marketId != "" {
		q.Set("marketId", marketId)
	}
	if resolution != "" {
		q.Set("resolution", resolution)
	}
	if countback > 0 {
		q.Set("countback", fmt.Sprintf("%d", countback))
	}
	// if countback is 0, set it to empty string means all data
	if countback == 0 {
		q.Set("countback", "")
	}
	if from != 0 {
		q.Set("from", fmt.Sprintf("%d", from))
	}
	if to != 0 {
		q.Set("to", fmt.Sprintf("%d", to))
	}

	endpoint := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotHistoryPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		logx.Errorf("SpotMarketHistory new request error: %v", err)
		return model.SpotMarketHistory{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logx.Errorf("SpotMarketHistory do request error: %v", err)
		return model.SpotMarketHistory{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logx.Errorf("SpotMarketHistory request response error: %v", fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b)))
		return model.SpotMarketHistory{}, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out model.SpotMarketHistory
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logx.Errorf("SpotMarketHistory decode response error: %v", err)
		return model.SpotMarketHistory{}, err
	}
	return out, nil
}

func (c *Client) SpotSymbolInfo(ctx context.Context, group string) (*model.SpotSymbolInfo, error) {
	q := url.Values{}
	q.Set("group", group)

	endpoint := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotSymbolInfoPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		logx.Errorf("SpotSymbolInfo new request error: %v", err)
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logx.Errorf("SpotSymbolInfo do request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logx.Errorf("SpotSymbolInfo request response error: %v", fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b)))
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out model.SpotSymbolInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logx.Errorf("SpotSymbolInfo decode response error: %v", err)
		return nil, err
	}
	return &out, nil
}

func (c *Client) SpotSymbols(ctx context.Context, symbol string) (*model.SpotSymbolsRaw, error) {
	q := url.Values{}
	if symbol != "" {
		q.Set("symbol", symbol)
	} else {
		return nil, fmt.Errorf("SpotSymbols request symbol is required")
	}
	endpoint := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.SpotSymbolsPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		logx.Errorf("SpotSymbols new request error: %v", err)
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logx.Errorf("SpotSymbols do request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logx.Errorf("SpotSymbols request response error: %v", fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b)))
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out model.SpotSymbolsRaw
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logx.Errorf("SpotSymbols decode response error: %v", err)
		return nil, err
	}
	return &out, nil
}
