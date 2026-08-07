package service

import (
	"context"
	"fmt"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type roleService struct {
	txManager model.TransactionManager
	roleRepo  model.RoleRepository
}

func NewRoleService(txManager model.TransactionManager, roleRepo model.RoleRepository) model.RoleService {
	return &roleService{txManager: txManager, roleRepo: roleRepo}
}

func (s *roleService) Access(ctx context.Context, acc *model.Access) (bool, error) {
	hasAccess, err := s.roleRepo.RolePermission(ctx, acc.UserId, acc.MenuName, acc.ActionName)
	if err != nil {
		return false, err
	}

	return hasAccess, nil
}

func (s *roleService) UpsertRole(ctx context.Context, upsertRole *model.UpsertRole) error {
	var menuIds []int
	var actionIds []int
	var err error
	for _, p := range upsertRole.RolePermissions {
		menuIds = append(menuIds, p.MenuId)
		actionIds = append(actionIds, p.ActionId)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("roleService=UpsertRole: %w", err)
	}

	defer tx.Rollback(ctx)

	if upsertRole.RoleId == 0 {
		err = s.roleRepo.CreateRole(tx.Context(), &upsertRole.Role)
	} else {
		err = s.roleRepo.UpdateRole(tx.Context(), &upsertRole.Role)
	}

	if err != nil {
		return fmt.Errorf("roleService=UpsertRole: %w", err)
	}

	if len(menuIds) == 0 {
		// Edge Case: The admin unchecked all boxes. Just wipe them all.
		err = s.roleRepo.DeleteAllRolePermissions(tx.Context(), upsertRole.RoleId)
		if err != nil {
			return fmt.Errorf("roleService=UpsertRole: %w", err)
		}
	} else {
		// Normal Case: Surgically delete removed ones, then insert new ones
		err = s.roleRepo.DeleteRolePermission(tx.Context(), upsertRole.RoleId, menuIds, actionIds)
		if err != nil {
			return fmt.Errorf("roleService=UpsertRole: %w", err)
		}

		err = s.roleRepo.CreateRolePermission(tx.Context(), upsertRole.RoleId, menuIds, actionIds)
		if err != nil {
			return fmt.Errorf("roleService=UpsertRole: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("roleService=UpsertRole: %w", err)
	}

	return nil
}

func (s *roleService) GetAll(ctx context.Context) ([]model.Role, error) {
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("roleService=GetAll: %w", err)
	}

	return roles, nil

}

func (s *roleService) GetMenuActionAvailable(ctx context.Context) ([]model.MainMenu, error) {
	mainMenus, err := s.roleRepo.GetMenuActionAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("roleService=GetMenuActionAvailable: %w", err)
	}

	return mainMenus, nil
}

func (s *roleService) GetDetailById(ctx context.Context, roleId int) (*model.RoleDetail, error) {
	detail, err := s.roleRepo.GetPermissionById(ctx, roleId)
	if err != nil {
		return nil, fmt.Errorf("roleService=GetMenuActionAvailable: %w", err)
	}

	return detail, nil
}
