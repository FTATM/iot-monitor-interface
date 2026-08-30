package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/auth"
	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

type userService struct {
	txManager    model.TransactionManager
	prefixError  string
	userRepo     model.UserRepository
	jwtKey       []byte
	roleRepo     model.RoleRepository
	auditLogRepo model.AuditLogRepository
}

func NewUserService(txManager model.TransactionManager, userRepo model.UserRepository, jwtKey []byte, roleRepo model.RoleRepository, auditLogRepo model.AuditLogRepository) model.UserService {
	return &userService{txManager: txManager, prefixError: "userService", userRepo: userRepo, jwtKey: jwtKey, roleRepo: roleRepo, auditLogRepo: auditLogRepo}
}

func (s *userService) CreateUser(ctx context.Context, createUser *model.CreateUser, authUserId int) (*model.User, error) {
	const fname = "CreateUser"
	hash, err := argon2id.CreateHash(createUser.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	user := model.User{
		FirstName:    createUser.FirstName,
		LastName:     createUser.LastName,
		Username:     createUser.Username,
		Active:       createUser.Active,
		PasswordHash: hash,
		RoleId:       createUser.RoleId,
		Email:        createUser.Email,
		Tel:          createUser.Tel,
	}

	countValidate, err := s.userRepo.CountValidate(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if countValidate > 0 {
		return nil, fmt.Errorf("[%s]>[%s]: user duplicate", s.prefixError, fname)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	newData, err := model.StructToDynamicJSON(user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(user.UserId),
		Action:     model.CreateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &user, nil
}

func (s *userService) UpdateUser(ctx context.Context, updateUser *model.UpdateUser, authUserId int) (*model.User, error) {
	const fname = "UpdateUser"
	var err error
	var hash string
	if len(updateUser.Password) != 0 {
		hash, err = argon2id.CreateHash(updateUser.Password, argon2id.DefaultParams)
		if err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	user := model.User{
		UserId:       updateUser.UserId,
		FirstName:    updateUser.FirstName,
		LastName:     updateUser.LastName,
		Active:       updateUser.Active,
		PasswordHash: hash,
		RoleId:       updateUser.RoleId,
		Email:        updateUser.Email,
		Tel:          updateUser.Tel,
	}

	countValidate, err := s.userRepo.CountValidate(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if countValidate > 0 {
		return nil, fmt.Errorf("[%s]>[%s]: user duplicate", s.prefixError, fname)
	}

	oldUser, err := s.userRepo.GetById(ctx, user.UserId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if oldUser.IsSame(user) && len(user.PasswordHash) == 0 {
		return &user, nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.Update(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldUser)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(user.UserId),
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &user, nil
}

func (s *userService) GetPermissionMapByUserId(ctx context.Context, userId int) (map[string][]string, error) {
	const fname = "GetPermissionMapByUserId"
	userActive, err := s.userRepo.GetActiveById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if !userActive {
		return nil, nil
	}

	rolePermissionDescs, err := s.roleRepo.GetPermissionDescByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	permissionMap := make(map[string][]string)
	for _, p := range rolePermissionDescs {
		permissionMap[p.MenuName] = append(permissionMap[p.MenuName], p.ActionName)
	}

	return permissionMap, nil
}

func (s *userService) LoginUserJwt(ctx context.Context, creds *model.LoginCredentials, issueTime, expTime time.Time) (*model.User, string, error) {
	const fname = "LoginUserJwt"
	var tokenString string
	user, err := s.userRepo.GetByUsername(ctx, creds.Username)
	if err != nil {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if !user.Active {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrNotActive)
	}

	match, err := argon2id.ComparePasswordAndHash(creds.Password, user.PasswordHash)
	if err != nil {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if !match {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: Invalid username or password", s.prefixError, fname)
	}

	claims := &auth.Claim{
		UserId:    user.UserId,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		// Permissions: permissionMap,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(issueTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err = token.SignedString(s.jwtKey)
	if err != nil {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return user, tokenString, nil
}

func (s *userService) GetAllDetail(ctx context.Context, active bool) ([]model.UserDetail, error) {
	const fname = "GetAllDetail"
	users, err := s.userRepo.GetAll(ctx, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	userDetails := make([]model.UserDetail, 0, len(users))

	for _, u := range users {
		detail := model.UserDetail{
			User: u,
		}

		userDetails = append(userDetails, detail)
	}
	return userDetails, nil
}

func (s *userService) DeleteUser(ctx context.Context, deleteUserId, authUserId int) error {
	const fname = "DeleteUser"
	var err error

	oldUser, err := s.userRepo.GetById(ctx, deleteUserId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.Delete(ctx, deleteUserId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldUser)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(deleteUserId),
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
