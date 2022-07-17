package chronos_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/xruins/chronos/lib/chronos"
)

type mockRoundTripper struct {
	response []byte
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet && req.URL.Path == chronos.HealthCheckEndpoint {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(m.response)),
		}, nil
	}
	return &http.Response{StatusCode: http.StatusNotFound}, nil
}

func newMockClient(response []byte) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{response: response},
	}
}

func TestCheckHealth(t *testing.T) {
	type pattern struct {
		description  string
		mockResponse string
		want         bool
		wantErr      bool
	}

	patterns := []*pattern{
		&pattern{
			description:  "returns true on healthy state",
			mockResponse: `{"ok":true}`,
			want:         true,
			wantErr:      false,
		},
		&pattern{
			description:  "returns false on unhealthy state",
			mockResponse: `{"ok":false,failed_jobs:["foo","bar"]}`,
			want:         false,
			wantErr:      true,
		},
	}

	for _, p := range patterns {
		httpClient := newMockClient([]byte(p.mockResponse))
		client := chronos.NewClient(httpClient)

		url := &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   chronos.HealthCheckEndpoint,
		}
		got, gotErr := client.CheckHealth(context.Background(), url)
		if got != p.want {
			t.Errorf("unexpected return value. got: %v, want: %v", got, p.want)
		}
		if (gotErr != nil) != p.wantErr {
			t.Errorf("unexpected return error. got: %s, want: %v", gotErr, p.wantErr)
		}
	}
}
