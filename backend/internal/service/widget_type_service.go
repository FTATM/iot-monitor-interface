package service

import (
	"context"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type widgetTypeService struct {
	txManager      model.TransactionManager
	widgetTypeRepo model.WidgetTypeRepository
}

func NewWidgetTypeService(txManager model.TransactionManager, wrt model.WidgetTypeRepository) model.WidgetTypeService {
	return &widgetTypeService{txManager: txManager, widgetTypeRepo: wrt}
}

func (s *widgetTypeService) GetWidgetTypeById(ctx context.Context, canvasId int) (*model.WidgetType, error) {

	widgetType, err := s.widgetTypeRepo.GetById(ctx, canvasId)
	if err != nil {
		return nil, err
	}

	return widgetType, nil
}

func (s *widgetTypeService) GetWidgetTypeAll(ctx context.Context) ([]model.WidgetType, error) {

	widgetTypes, err := s.widgetTypeRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return widgetTypes, nil
}
