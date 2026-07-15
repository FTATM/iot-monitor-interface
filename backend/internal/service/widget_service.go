package service

import (
	"context"
	"fmt"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type widgetService struct {
	txManager      model.TransactionManager
	widgetRepo     model.WidgetRepository
	widgetTypeRepo model.WidgetTypeRepository
	canvasRepo     model.CanvasRepository
}

func NewWidgetService(txManager model.TransactionManager, wr model.WidgetRepository, wrt model.WidgetTypeRepository, cr model.CanvasRepository) model.WidgetService {
	return &widgetService{
		txManager:      txManager,
		widgetRepo:     wr,
		widgetTypeRepo: wrt,
		canvasRepo:     cr,
	}
}

func (s *widgetService) GetWidgetDetailById(ctx context.Context, widgetId int) (*model.Widget, error) {
	widget, err := s.widgetRepo.GetById(ctx, widgetId)
	if err != nil {
		return nil, err
	}

	return &model.Widget{
		WidgetId:     widget.WidgetId,
		WidgetTypeId: widget.WidgetTypeId,
		LayoutData:   widget.LayoutData,
	}, nil
}

func (s *widgetService) CreateWidgets(ctx context.Context, widgetRequest []model.Widget) error {
	var err error
	if len(widgetRequest) == 0 {
		return fmt.Errorf("no widget request")
	}

	_, err = s.canvasRepo.GetById(ctx, widgetRequest[0].CanvasId)
	if err != nil {
		return fmt.Errorf("invalid canvas ID provided: %d", widgetRequest[0].CanvasId)
	}

	widgetTypes, err := s.widgetTypeRepo.GetAll(ctx)
	widgetTypeMap := make(map[int]string)

	for _, widgetType := range widgetTypes {
		widgetTypeMap[widgetType.WidgetTypeId] = widgetType.WidgetTypeName
	}

	for _, widget := range widgetRequest {
		_, exists := widgetTypeMap[widget.WidgetTypeId]

		if !exists {
			return fmt.Errorf("invalid widget type ID provided: %d", widget.WidgetTypeId)
		}

	}

	if err != nil {
		return err
	}

	if err = s.widgetRepo.Create(ctx, widgetRequest); err != nil {
		return err
	}

	return nil
}

func (s *widgetService) UpdateWidget(ctx context.Context, widgetRequest []model.Widget) error {
	var err error
	if len(widgetRequest) == 0 {
		return fmt.Errorf("no widget request")
	}

	if err = s.widgetRepo.Update(ctx, widgetRequest); err != nil {
		return err
	}

	return nil
}

func (s *widgetService) DeleteWidgets(ctx context.Context, widgetIds []int) error {
	var err error
	if err = s.widgetRepo.Delete(ctx, widgetIds); err != nil {
		return err
	}

	return nil
}
