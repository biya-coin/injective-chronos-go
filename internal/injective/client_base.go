package injective

import (
	"net/http"

	"github.com/biya-coin/injective-chronos-go/internal/config"
)

type Client struct {
	cfg        config.InjectiveConf
	httpClient *http.Client
}

func NewClient(cfg config.InjectiveConf, hc *http.Client) *Client {
	return &Client{cfg: cfg, httpClient: hc}
}
