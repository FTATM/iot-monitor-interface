package router

import (
	"github.com/FTATM/iot-monitor-interface/internal/handler"
)

type RouterHandlers struct {
	Widget         *handler.WidgetHandler
	Canvas         *handler.CanvasHandler
	WidgetType     *handler.WidgetTypeHandler
	User           *handler.UserHandler
	Device         *handler.DeviceHandler
	Schedule       *handler.ScheduleHandler
	ScheduleEngine *handler.ScheduleEngineHandler
	Role           *handler.RoleHandler
	DeviceGateway  *handler.DeviceGatewayHandler
	// Add more as the app grows...
}
