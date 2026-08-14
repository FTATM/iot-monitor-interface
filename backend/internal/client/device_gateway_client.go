package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

// GatewayResponse represents the JSON structure returned by the Gateway
type GatewayResponse struct {
	Message string `json:"message"`
	Data    struct {
		DeviceId int  `json:"deviceId"`
		IsOnline bool `json:"isOnline"`
	} `json:"data"`
}

type gatewayClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewDeviceGatewayClient initializes the client with a strict timeout for internal calls
func NewDeviceGatewayClient(baseURL string) model.DeviceGatewayClient {
	return &gatewayClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // Prevents the API from hanging if the Gateway crashes
		},
	}
}

// GetDeviceStatus makes the HTTP GET request to the Gateway
func (c *gatewayClient) GetDeviceStatus(ctx context.Context, deviceId int) (bool, error) {
	url := fmt.Sprintf("%s/devicegateway/internal/devicestatus?deviceId=%d", c.baseURL, deviceId)

	// Use NewRequestWithContext so the request cancels if the user closes their browser
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("gateway returned unexpected status: %d", resp.StatusCode)
	}

	var gatewayData GatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gatewayData); err != nil {
		return false, fmt.Errorf("failed to decode gateway response: %w", err)
	}

	return gatewayData.Data.IsOnline, nil
}

func (c *gatewayClient) ExecuteManualCommand(ctx context.Context, req *model.CommandRequest) error {
	url := fmt.Sprintf("%s/api/v1/commands/manual", c.baseURL)

	// Encode the request body
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to encode command: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gateway returned unexpected status: %d", resp.StatusCode)
	}

	return nil
}
