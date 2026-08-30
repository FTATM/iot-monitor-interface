package listener

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func StartMQTTClient(ctx context.Context, brokerURL string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("device-gateway-worker")
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Second)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.InfoContext(ctx, "Connected to MQTT Broker!")

		// Subscribe to Single Device Topic
		singleTopic := "device/+/data"
		sHandler := singleHandler(ctx, svc, sessionSvc)
		if token := c.Subscribe(singleTopic, 1, sHandler); token.Wait() && token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to subscribe to MQTT topic", slog.String("error", token.Error().Error()))
		} else {
			slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", singleTopic))
		}

		// Subscribe to Group Device Topic
		groupTopic := "device-group/+/data"
		gHandler := groupHandler(ctx, svc, sessionSvc)
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
func singleHandler(ctx context.Context, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topicParts := strings.Split(msg.Topic(), "/")
		if len(topicParts) != 3 {
			return
		}
		deviceName := topicParts[1]

		// 1. Resolve Name to ID using TTL Cache
		var deviceId int
		if cachedData, ok := getFromCache(&deviceNameCache, deviceName); ok {
			deviceId = cachedData.(int)
		} else {
			dbId, err := svc.GetDeviceIdByName(ctx, deviceName)
			if err != nil || dbId <= 0 {
				slog.Warn("Unknown device name received", slog.String("name", deviceName))
				return
			}
			setToCache(&deviceNameCache, deviceName, dbId, cacheTTL)
			deviceId = dbId
		}

		payload := bytes.TrimSpace(msg.Payload())
		if len(payload) == 0 {
			return
		}

		var incomingData []model.DeviceDataPayloadReq
		switch payload[0] {
		case '[':
			if err := json.Unmarshal(payload, &incomingData); err != nil {
				return
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(payload, &single); err != nil {
				return
			}
			incomingData = append(incomingData, single)
		default:
			return
		}

		// Process incoming array
		for _, d := range incomingData {
			deviceData := model.DeviceData{
				DeviceId:  deviceId,
				ValueData: d.ValueData,
			}
			sessionSvc.MarkDeviceActive(deviceData.DeviceId)
			svc.Add(deviceData)
		}
	}
}

func groupHandler(ctx context.Context, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topicParts := strings.Split(msg.Topic(), "/")
		if len(topicParts) != 3 {
			return
		}
		groupName := topicParts[1]

		// 1. Resolve Group Name to Device Names using TTL Cache
		var deviceNames []string
		if cachedData, ok := getFromCache(&deviceGroupNameCache, groupName); ok {
			deviceNames = cachedData.([]string)
		} else {
			groupDataList, err := svc.GetDeviceIdByGroupName(ctx, groupName)
			if err != nil || len(groupDataList) == 0 {
				slog.Warn("Unknown or empty group name received", slog.String("groupName", groupName))
				return
			}
			deviceNames = groupDataList[0].DeviceNames
			setToCache(&deviceGroupNameCache, groupName, deviceNames, cacheTTL)
		}

		payload := bytes.TrimSpace(msg.Payload())
		if len(payload) == 0 {
			return
		}

		var incomingData []model.DeviceDataPayloadReq
		switch payload[0] {
		case '[':
			if err := json.Unmarshal(payload, &incomingData); err != nil {
				return
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(payload, &single); err != nil {
				return
			}
			incomingData = append(incomingData, single)
		default:
			return
		}

		// 2. FAN-OUT: Loop through EVERY device in the group
		for _, dName := range deviceNames {

			// Resolve Device ID using TTL Cache
			var deviceId int
			if cachedData, ok := getFromCache(&deviceNameCache, dName); ok {
				deviceId = cachedData.(int)
			} else {
				dbId, err := svc.GetDeviceIdByName(ctx, dName)
				if err != nil || dbId <= 0 {
					slog.Warn("Unknown device in group", slog.String("device", dName), slog.String("group", groupName))
					continue
				}
				setToCache(&deviceNameCache, dName, dbId, cacheTTL)
				deviceId = dbId
			}

			// Process payload for this specific device
			for _, d := range incomingData {
				deviceData := model.DeviceData{
					DeviceId:  deviceId,
					ValueData: d.ValueData,
				}
				sessionSvc.MarkDeviceActive(deviceData.DeviceId)
				svc.Add(deviceData)
			}
		}
	}
}
