package chronos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client it an HTTP Client for HealthCheck.
type Client struct {
	httpClient *http.Client
}

// NewClient returns an instance of Client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
	}
}

func (c *Client) getClient() *http.Client {
	if c.httpClient == nil {
		return http.DefaultClient
	}
	return c.httpClient
}

// CheckHealth invokes HealthCheck API and return the status of Chronos worker.
// It returns `true` if Chronos worker in healthy state.
func (c *Client) CheckHealth(ctx context.Context, u *url.URL) (bool, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: map[string][]string{
			"Accept": {"application/json"},
		},
	}

	req = req.WithContext(ctx)
	res, err := c.getClient().Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to exec a request for healthCheckAPI: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read the body of response: %w", err)
	}
	healthCheckResult := &healthCheckResult{}
	err = json.Unmarshal(b, healthCheckResult)
	if err != nil {
		return false, fmt.Errorf("malformed response: %w", err)
	}

	return healthCheckResult.OK, nil
}
