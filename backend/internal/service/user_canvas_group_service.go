package service

import (
	"context"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type userCanvasGroupService struct {
	txManager           model.TransactionManager
	userCanvasGroupRepo model.UserCanvasGroupRepository
}

func NewuserCanvasGroupService(txManager model.TransactionManager, ucg model.UserCanvasGroupRepository) model.UserCanvasGroupService {
	return &userCanvasGroupService{txManager: txManager, userCanvasGroupRepo: ucg}
}

func (s *userCanvasGroupService) GetAllCanvasByUserId(ctx context.Context, id int) ([]model.UserCanvasGroup, error) {

	userCanvasGroup, err := s.userCanvasGroupRepo.GetByUserId(ctx, id)
	if err != nil {
		return nil, err
	}

	return userCanvasGroup, nil
}
