package edgenetwork

import (
	"net/http"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
)

// Client is a lightweight Gcore API client for Edge Network products (CDN, DNS, FastEdge, WAAP).
type Client struct {
	RootURL string
	APIKey  string
	HTTP    *http.Client
}

func (c *Client) setHeaders(req *http.Request) {
	core.ApplyJSONAuthHeaders(req, c.APIKey)
}

