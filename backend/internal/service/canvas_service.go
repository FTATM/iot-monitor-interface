package service

import (
	"context"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type canvasService struct {
	txManager                 model.TransactionManager
	canvasRepo                model.CanvasRepository
	widgetRepo                model.WidgetRepository
	userCanvasGroupRepository model.UserCanvasGroupRepository
}

func NewCanvasService(txManager model.TransactionManager, wr model.WidgetRepository, cr model.CanvasRepository, ucg model.UserCanvasGroupRepository) model.CanvasService {
	return &canvasService{txManager: txManager, widgetRepo: wr, canvasRepo: cr, userCanvasGroupRepository: ucg}
}

func (s *canvasService) GetCanvasDetailById(ctx context.Context, canvasId int) (*model.CanvasDetail, error) {

	canvas, err := s.canvasRepo.GetById(ctx, canvasId)
	if err != nil {
		return nil, err
	}

	widgets, err := s.widgetRepo.GetWidgetByCanvasId(ctx, canvas.CanvasId)

	return &model.CanvasDetail{
		CanvasId: canvas.CanvasId,
		Widgets:  widgets,
	}, nil
}

func (s *canvasService) GetAllCanvasByUserId(ctx context.Context, id int) ([]model.CanvasDetail, error) {

	userCanvas, err := s.userCanvasGroupRepository.GetCanvasByUserId(ctx, id)
	// canvasDetailResult := make([]model.CanvasDetail)
	if err != nil {
		return nil, err
	}

	canvasIds := make([]int, 0, len(userCanvas))
	for _, canvas := range userCanvas {
		canvasIds = append(canvasIds, canvas.CanvasId)
	}

	widgets, err := s.widgetRepo.GetWidgetByCanvasIds(ctx, canvasIds)
	if err != nil {
		return nil, err
	}

	widgetMap := make(map[int][]model.Widget, len(userCanvas))
	for _, widget := range widgets {
		widgetMap[widget.CanvasId] = append(widgetMap[widget.CanvasId], widget)
	}

	result := make([]model.CanvasDetail, 0, len(userCanvas))

	for _, canvas := range userCanvas {
		detail := model.CanvasDetail{
			CanvasId:   canvas.CanvasId,
			CanvasName: canvas.CanvasName,
			Widgets:    widgetMap[canvas.CanvasId],
		}
		result = append(result, detail)
	}

	return result, nil
}
