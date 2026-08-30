package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type notificationClient struct {
	sms         config.Sms
	email       config.Email
	line        config.Line
	httpClient  *http.Client
	prefixError string
}

func NewNotificationClient(sms config.Sms, email config.Email, line config.Line) model.NotificationClient {
	return &notificationClient{
		sms:   sms,
		email: email,
		line:  line,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		prefixError: "notificationClient",
	}
}

func (n *notificationClient) SendSms(ctx context.Context, smsUser []model.UserNotificationSend) error {
	const fname = "SendSms"

	for _, u := range smsUser {
		tel := strings.TrimSpace(u.Tel)
		if after, ok := strings.CutPrefix(tel, "0"); ok {
			tel = "66" + after
		}
		payload := struct {
			Msisdn  string `json:"msisdn"`
			Message string `json:"message"`
			Sender  string `json:"sender"`
			Force   string `json:"force"`
		}{
			Msisdn:  tel,
			Message: u.Msg,
			Sender:  n.sms.Sender,
			Force:   "standard",
		}

		payloadJsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", n.prefixError, fname, err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.sms.Url, bytes.NewReader(payloadJsonData))
		if err != nil {
			return fmt.Errorf("[%s]>[%s] failed to create request: %w", n.prefixError, fname, err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(n.sms.Key, n.sms.Secret)

		resp, err := n.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("[%s]>[%s] request failed: %w", n.prefixError, fname, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("[%s]>[%s] returned unexpected status: %w", n.prefixError, fname, err)
		}
	}

	return nil
}
