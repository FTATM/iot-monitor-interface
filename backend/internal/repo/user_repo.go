package repo

import (
	"context"
	"errors"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool DBTX
}

// NewUserRepository creates a new repository instance
func NewUserRepository(pool *pgxpool.Pool) model.UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx // Found a transaction in the context!
	}
	return r.pool // No transaction, use standard pool
}

func (r *userRepo) GetById(ctx context.Context, id int) (*model.User, error) {
	user := &model.User{}
	query := `SELECT user_id, first_name, last_name, active FROM "user" WHERE user_id = $1`
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(
		&user.UserId,
		&user.FirstName,
		&user.LastName,
		&user.Active,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepo) UserCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT count(*) FROM "user"`
	err := r.db(ctx).QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `SELECT user_id, first_name, last_name, password_hash, active FROM "user" WHERE username = $1`
	err := r.db(ctx).QueryRow(ctx, query, username).Scan(
		&user.UserId,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.Active,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("Username not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	query := `
			INSERT INTO "user" (first_name, last_name, username, password_hash, active)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING user_id
		`
	err := r.db(ctx).QueryRow(ctx, query, &user.FirstName, &user.LastName, &user.Username, &user.PasswordHash, &user.Active).Scan(&user.UserId)

	if err != nil {
		return errors.New("Insert user fail")
	}

	return nil

}

// func (r *userRepo) GetByIds(ctx context.Context, ids []int) ([]model.Canvas, error) {
// 	query := "SELECT canvas_id, canvas_name FROM canvas WHERE canvas_id = ANY($1)"
// 	rows, err := r.db(ctx).Query(ctx, query, ids)
// 	if err != nil {
// 		return nil, err
// 	}

// 	canvasList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
// 	if err != nil {
// 		return nil, err
// 	}

// 	return canvasList, nil
// }
