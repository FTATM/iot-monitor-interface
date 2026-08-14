package listener

import (
	"context"
	"encoding/json"
	"log/slog"
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

	// Define what happens when a message arrives
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		var data model.DeviceDataRequest
		// 1. Validate JSON syntax
		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			slog.Debug("Invalid JSON syntax received over MQTT",
				slog.String("error", err.Error()),
				slog.String("topic", msg.Topic()),
				slog.String("payload", string(msg.Payload())),
			)
			return
		}

		// 2. Validate required struct fields (Ensure DeviceId is valid)
		if data.DeviceId <= 0 {
			slog.Debug("MQTT message rejected: invalid or missing DeviceId",
				slog.String("topic", msg.Topic()),
				slog.String("payload", string(msg.Payload())),
			)
			return
		}

		sessionSvc.MarkDeviceActive(data.DeviceId)
		// Pass to the Batcher Service!
		svc.Add(data)
	})

	// opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
	// 	var data map[string]int
	// 	// 2. Validate required struct fields (Ensure DeviceId is valid)
	// 	// if data.DeviceId <= 0 {
	// 	// 	slog.Debug("MQTT message rejected: invalid or missing DeviceId",
	// 	// 		slog.String("topic", msg.Topic()),
	// 	// 		slog.String("payload", string(msg.Payload())),
	// 	// 	)
	// 	// 	return
	// 	// }

	// 	// sessionSvc.MarkDeviceActive(data.DeviceId)
	// 	// Pass to the Batcher Service!
	// 	err := json.NewDecoder(msg.Payload()).Decode(&data)
	// 	if err != nil {
	// 		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// 	defer r.Body.Close()
	// 	var device model.DeviceDataRequest
	// 	device.DeviceName = data
	// 	svc.Add(data)
	// })

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.InfoContext(ctx, "Connected to MQTT Broker!")
		// Subscribe to a wildcard topic where devices publish data
		topic := "devices/+/data"
		if token := c.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to subscribe to MQTT topic", slog.String("error", token.Error().Error()))
		} else {
			slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", topic))
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
