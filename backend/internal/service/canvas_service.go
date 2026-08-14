package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type canvasService struct {
	txManager    model.TransactionManager
	prefixError  string
	canvasRepo   model.CanvasRepository
	widgetRepo   model.WidgetRepository
	auditLogRepo model.AuditLogRepository
}

func NewCanvasService(txManager model.TransactionManager, wr model.WidgetRepository, cr model.CanvasRepository, auditlogRepo model.AuditLogRepository) model.CanvasService {
	return &canvasService{txManager: txManager, prefixError: "canvasService", widgetRepo: wr, canvasRepo: cr, auditLogRepo: auditlogRepo}
}

func (s *canvasService) GetAllCanvas(ctx context.Context) ([]model.Canvas, error) {
	const fname = "GetAllCanvasByUserRole"

	allCanvas, err := s.canvasRepo.GetAll(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return allCanvas, nil
}

func (s *canvasService) GetCanvasDetailById(ctx context.Context, canvasId int) (*model.CanvasDetail, error) {
	const fname = "GetCanvasDetailById"

	canvas, err := s.canvasRepo.GetById(ctx, canvasId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	widgets, err := s.widgetRepo.GetWidgetByCanvasId(ctx, []int{canvas.CanvasId})
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &model.CanvasDetail{
		CanvasId: canvas.CanvasId,
		Widgets:  widgets,
	}, nil
}

func (s *canvasService) GetAllCanvasDetailByUserRole(ctx context.Context, authUserId int) ([]model.CanvasDetail, error) {
	const fname = "GetAllCanvasDetailByUserRole"

	userCanvas, err := s.canvasRepo.GetCanvasByUserId(ctx, authUserId, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	canvasIds := make([]int, 0, len(userCanvas))
	for _, canvas := range userCanvas {
		canvasIds = append(canvasIds, canvas.CanvasId)
	}

	widgets, err := s.widgetRepo.GetWidgetByCanvasId(ctx, canvasIds)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
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

func (s *canvasService) GetAllCanvasRoleDetail(ctx context.Context) ([]model.CanvasRoleDetail, error) {
	const fname = "GetAllCanvasRoleDetail"

	allCanvasRole, err := s.canvasRepo.GetAllCanvasRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	roleMap := make(map[int][]int)
	for _, cr := range allCanvasRole {
		roleMap[cr.RoleId] = append(roleMap[cr.RoleId], cr.CanvasId)
	}

	var detail []model.CanvasRoleDetail
	for roleId, canvasIds := range roleMap {
		detail = append(detail, model.CanvasRoleDetail{
			RoleId:    roleId,
			CanvasIds: canvasIds,
		})
	}

	return detail, nil
}

func (s *canvasService) UpsertCanvasRole(ctx context.Context, upsertCanvasRole *model.UpsertCanvasRole, authUserId int) error {
	const fname = "UpsertCanvasRole"
	var err error

	oldCanvasIds, err := s.canvasRepo.GetCanvasRoleByRoleId(ctx, upsertCanvasRole.RoleId)
	oldCanvasMap := make(map[int]bool)
	for _, oc := range oldCanvasIds {
		oldCanvasMap[oc] = true
	}

	var createCanvasRole, deleteCanvasRole []model.CanvasRole

	for _, cId := range upsertCanvasRole.CanvasIds {
		if _, found := oldCanvasMap[cId]; found {
			// Found in both!
			// Delete it from the map so we know it's been handled.
			delete(oldCanvasMap, cId)
		} else {
			//Not found in old map! This is a brand new ID.
			createCanvasRole = append(createCanvasRole, model.CanvasRole{RoleId: upsertCanvasRole.RoleId, CanvasId: cId})
		}
	}

	for cId := range oldCanvasMap {
		deleteCanvasRole = append(deleteCanvasRole, model.CanvasRole{RoleId: upsertCanvasRole.RoleId, CanvasId: cId})
	}

	if len(createCanvasRole) == 0 && len(deleteCanvasRole) == 0 {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.CreateCanvasRole(ctx, createCanvasRole)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	err = s.canvasRepo.DeleteCanvasRole(ctx, deleteCanvasRole)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// auditlogs := make([]model.AuditLog, 0, 1)
	// oldData, err := model.StructToDynamicJSON(oldUser)
	// if err != nil {
	// 	return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	// }
	// newData, err := model.StructToDynamicJSON(user)
	// if err != nil {
	// 	return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	// }
	// audit := model.AuditLog{
	// 	EntityType: "canvas_role",
	// 	EntityId:   strconv.Itoa(user.UserId),
	// 	Action:     model.UpdateAction,
	// 	ChangedBy:  authUserId,
	// 	OldData:    oldData,
	// 	NewData:    newData,
	// }
	// auditlogs = append(auditlogs, audit)

	// if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
	// 	return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	// }

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *canvasService) CreateCanvas(ctx context.Context, createCanvas *model.CreateCanvas, authUserId int) error {
	const fname = "CreateCanvas"

	canvas := model.Canvas{
		CanvasName: createCanvas.CanvasName,
	}

	duplicate, err := s.canvasRepo.CountValidate(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if duplicate > 0 {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrDuplicate)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.Create(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	newData, err := model.StructToDynamicJSON(canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(canvas.CanvasId),
		Action:     model.CreateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
func (s *canvasService) UpdateCanvas(ctx context.Context, updateCanvas *model.UpdateCanvas, authUserId int) error {
	const fname = "UpdateCanvas"

	canvas := model.Canvas{
		CanvasId:   updateCanvas.CanvasId,
		CanvasName: updateCanvas.CanvasName,
	}

	duplicate, err := s.canvasRepo.CountValidate(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if duplicate > 0 {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrDuplicate)
	}

	oldCanvas, err := s.canvasRepo.GetById(ctx, canvas.CanvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if oldCanvas.IsSame(canvas) {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.Update(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldCanvas)
	newData, err := model.StructToDynamicJSON(canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(canvas.CanvasId),
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *canvasService) DeleteCanvas(ctx context.Context, canvasId int, authUserId int) error {
	const fname = "DeleteCanvas"

	oldCanvas, err := s.canvasRepo.GetById(ctx, canvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.Delete(ctx, canvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldCanvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(canvasId),
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
