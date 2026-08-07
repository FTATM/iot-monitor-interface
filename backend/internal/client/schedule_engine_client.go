package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type scheduleHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewSchedulerClient creates a new HTTP client with a strict 5-second timeout
func NewScheduleClient(baseURL string) model.ScheduleClient {
	return &scheduleHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *scheduleHTTPClient) SyncSchedule(scheduleID string) error {
	// Construct the URL: http://localhost:8081/api/schedules/{id}/sync
	url := fmt.Sprintf("%s/scheduleengine/%s/sync", c.baseURL, scheduleID)

	// Create the POST request (No body needed, just the URL)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create sync request: %w", err)
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("scheduler service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("scheduler service returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func (c *scheduleHTTPClient) UnsyncSchedule(scheduleID string) error {
	// Construct the URL: http://localhost:8081/api/schedules/{id}/sync
	url := fmt.Sprintf("%s/api/schedules/%s/sync", c.baseURL, scheduleID)

	// Create the DELETE request
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create unsync request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("scheduler service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("scheduler service returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
