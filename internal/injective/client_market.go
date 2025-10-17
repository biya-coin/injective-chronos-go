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

// MarketHistory fetches candle history for multiple marketIDs with resolution and countback.
// Example path: /api/chart/v1/market/history?marketIDs=...&marketIDs=...&resolution=5&countback=2
func (c *Client) MarketHistory(ctx context.Context, marketIDs []string, resolution string, countback int) ([]model.MarketHistory, error) {
	if len(marketIDs) == 0 {
		return nil, fmt.Errorf("MarketHistory request marketIDs is required")
	}
	if resolution == "" {
		return nil, fmt.Errorf("MarketHistory request resolution is required")
	}

	q := url.Values{}
	for _, id := range marketIDs {
		if id != "" {
			q.Add("marketIDs", id)
		}
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

	endpoint := fmt.Sprintf("%s%s", c.cfg.BaseURL, consts.MarketHistoryPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		logx.Errorf("MarketHistory new request error: %v", err)
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logx.Errorf("MarketHistory do request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logx.Errorf("MarketHistory request response error: %v", fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b)))
		return nil, fmt.Errorf("injective http %d: %s", resp.StatusCode, string(b))
	}
	var out []model.MarketHistory
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logx.Errorf("MarketHistory decode response error: %v", err)
		return nil, err
	}
	return out, nil
}
