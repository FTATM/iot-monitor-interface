package listener

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func StartMQTTClient(ctx context.Context, brokerURL string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, cacheSvc model.CacheService) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("device-gateway-worker")
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Second)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.InfoContext(ctx, "Connected to MQTT Broker!")

		// Subscribe to Single Device Topic
		singleTopic := "device/+/data"
		// ⚡ ส่ง cacheSvc เข้าไปใน Handler
		sHandler := singleHandler(ctx, svc, sessionSvc, cacheSvc)
		if token := c.Subscribe(singleTopic, 1, sHandler); token.Wait() && token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to subscribe to MQTT topic", slog.String("error", token.Error().Error()))
		} else {
			slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", singleTopic))
		}

		// Subscribe to Group Device Topic
		groupTopic := "device-group/+/data"
		// ⚡ ส่ง cacheSvc เข้าไปใน Handler
		gHandler := groupHandler(ctx, svc, sessionSvc, cacheSvc)
		if token := c.Subscribe(groupTopic, 1, gHandler); token.Wait() && token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to subscribe to group topic", slog.String("error", token.Error().Error()))
		} else {
			slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", groupTopic))
		}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.ErrorContext(ctx, "Failed to connect to MQTT broker", slog.String("error", token.Error().Error()))
		return
	}
	sessionSvc.SetMQTTClient(client)

	// Wait for graceful shutdown
	<-ctx.Done()
	slog.Info("Shutting down MQTT listener...")
	client.Disconnect(250)
}

// --- HANDLERS ---
func singleHandler(ctx context.Context, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, cacheSvc model.CacheService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topicParts := strings.Split(msg.Topic(), "/")
		if len(topicParts) != 3 {
			return
		}
		deviceName := topicParts[1]

		// 1. Resolve Name to ID using Central Cache Service
		deviceId, err := cacheSvc.GetDeviceIdByName(ctx, deviceName)
		if err != nil || deviceId <= 0 {
			slog.Warn("Unknown device name received", slog.String("name", deviceName))
			return
		}

		payload := bytes.TrimSpace(msg.Payload())
		if len(payload) == 0 {
			return
		}

		var incomingData []model.DeviceDataPayloadReq
		switch payload[0] {
		case '[':
			if err := json.Unmarshal(payload, &incomingData); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(payload, &single); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
			incomingData = append(incomingData, single)
		default:
			return
		}

		// Process payload for this specific device
		for _, d := range incomingData {
			deviceData := model.DeviceData{
				DeviceId:  deviceId,
				ValueData: int(math.Round(d.ValueData * model.DeviceScale)),
			}
			sessionSvc.MarkDeviceActive(deviceData.DeviceId)
			svc.Add(deviceData)
		}
	}
}

func groupHandler(ctx context.Context, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, cacheSvc model.CacheService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topicParts := strings.Split(msg.Topic(), "/")
		if len(topicParts) != 3 {
			return
		}
		groupName := topicParts[1]

		// 1. Resolve Group Name to Device Names using Central Cache Service
		deviceNames, err := cacheSvc.GetDeviceNamesByGroupName(ctx, groupName)
		if err != nil || len(deviceNames) == 0 {
			slog.Warn("Unknown or empty group name received", slog.String("groupName", groupName))
			return
		}

		payload := bytes.TrimSpace(msg.Payload())
		if len(payload) == 0 {
			return
		}

		var incomingData []model.DeviceDataPayloadReq
		switch payload[0] {
		case '[':
			if err := json.Unmarshal(payload, &incomingData); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(payload, &single); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
			incomingData = append(incomingData, single)
		default:
			return
		}

		// 2. Validate and process ONLY the specific devices sent in the payload
		validDevices := make(map[string]bool)
		for _, name := range deviceNames {
			validDevices[name] = true
		}

		// Loop through the incoming JSON payload array
		for _, d := range incomingData {
			dName := d.DeviceName

			if dName == "" || !validDevices[dName] {
				slog.Warn("Device in payload is missing or does not belong to group",
					slog.String("device", dName),
					slog.String("group", groupName),
				)
				continue
			}

			// Resolve Device ID using Central Cache Service
			deviceId, err := cacheSvc.GetDeviceIdByName(ctx, dName)
			if err != nil || deviceId <= 0 {
				slog.Warn("Unknown device in group", slog.String("device", dName))
				continue
			}

			deviceData := model.DeviceData{
				DeviceId:  deviceId,
				ValueData: int(math.Round(d.ValueData * model.DeviceScale)),
			}
			sessionSvc.MarkDeviceActive(deviceData.DeviceId)
			svc.Add(deviceData)
		}
	}
}
