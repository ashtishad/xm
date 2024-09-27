package domain

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ashtishad/xm/common"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, common.AppError)
	FindBy(ctx context.Context, dbColumnName string, value any) (*User, common.AppError)
}

type userRepository struct {
	db *sql.DB
	l  *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) UserRepository {
	return &userRepository{
		db: db,
		l:  logger,
	}
}

func (r userRepository) Create(ctx context.Context, user *User) (*User, common.AppError) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.l.Error(common.ErrTXBegin, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	defer r.rollBackOnError(tx)

	exists := false
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`, user.Email).Scan(&exists)
	if err != nil {
		r.l.Error("failed to check user existence", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if exists {
		return nil, common.NewConflictError("user with this email already exists")
	}

	queryCreateUser := `INSERT INTO users (uuid, email, name, password_hash, status, created_at, updated_at)
                        VALUES($1, $2, $3, $4, $5, $6, $7)
                        RETURNING id`

	var createdID int
	err = tx.QueryRowContext(ctx, queryCreateUser,
		user.UUID, user.Email, user.Name, user.PasswordHash, user.Status, user.CreatedAt, user.UpdatedAt).Scan(&createdID)

	if err != nil {
		r.l.Error("failed to create user", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		r.l.Error(common.ErrTxCommit, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	user.ID = createdID
	return user, nil
}

func (r userRepository) FindBy(ctx context.Context, dbColumnName string, value any) (*User, common.AppError) {
	query, err := generateFindByQuery(dbColumnName)
	if err != nil || query == "" {
		return nil, common.NewBadRequestError(common.ErrUnexpectedDatabase)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		r.l.Error(common.ErrTXBegin, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	defer r.rollBackOnError(tx)

	var user User
	err = tx.QueryRowContext(ctx, query, value).Scan(
		&user.ID, &user.UUID, &user.Email, &user.Name, &user.PasswordHash,
		&user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.NewNotFoundError("user not found")
		}

		r.l.Error("failed to get user", "field", dbColumnName, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		r.l.Error(common.ErrTxCommit, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return &user, nil
}

func (r userRepository) rollBackOnError(tx *sql.Tx) {
	if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
		r.l.Error(common.ErrTXRollback, "rbErr", rbErr)
	}
}

// generateFindByQuery is a helper for generating sql query for FindBy method.
// Allowed Fields: id, uuid, email
func generateFindByQuery(fieldName string) (string, error) {
	var query string

	switch fieldName {
	case common.DBColumnID:
		query = `SELECT id, uuid, email, name, password_hash, status, created_at, updated_at
                 FROM users WHERE id = $1`
	case common.DBColumnUUID:
		query = `SELECT id, uuid, email, name, password_hash, status, created_at, updated_at
                 FROM users WHERE uuid = $1`
	case common.DBColumnEmail:
		query = `SELECT id, uuid, email, name, password_hash, status, created_at, updated_at
                 FROM users WHERE email = $1`
	default:
		return "", errors.New("invalid db field name")
	}

	return query, nil
}
