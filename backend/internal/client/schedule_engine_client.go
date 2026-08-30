package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type scheduleEngineClient struct {
	baseURL        string
	httpClient     *http.Client
	internalSecret string
	prefixError    string
}

// NewSchedulerClient creates a new HTTP client with a strict 5-second timeout
func NewScheduleClient(baseURL, internalSecret string) model.ScheduleClient {
	return &scheduleEngineClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		internalSecret: internalSecret,
		prefixError:    "scheduleEngineClient",
	}
}

func (c *scheduleEngineClient) SyncSchedule(ctx context.Context, scheduleID string) error {
	const fname = "SyncSchedule"
	url := fmt.Sprintf("%s/scheduleengine/%s/sync", c.baseURL, scheduleID)

	// Create the POST request (No body needed, just the URL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq.Header.Set("X-Internal-Secret", c.internalSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] scheduler service unreachable: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s] scheduler service returned non-200 status: %d", c.prefixError, fname, resp.StatusCode)
	}

	return nil
}

func (c *scheduleEngineClient) UnsyncSchedule(ctx context.Context, scheduleID string) error {
	const fname = "UnsyncSchedule"
	url := fmt.Sprintf("%s/api/scheduleengine/%s/sync", c.baseURL, scheduleID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", c.prefixError, fname, err)
	}

	httpReq.Header.Set("X-Internal-Secret", c.internalSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] scheduler service unreachable: %w", c.prefixError, fname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%s]>[%s] scheduler service returned non-200 status: %d", c.prefixError, fname, resp.StatusCode)
	}

	return nil
}
