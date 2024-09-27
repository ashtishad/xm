package domain

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ashtishad/xm/common"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CompanyRepository defines the interface for company data operations.
type CompanyRepository interface {
	Create(ctx context.Context, company *Company) (*Company, common.AppError)
	FindByID(ctx context.Context, id uuid.UUID) (*Company, common.AppError)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*Company, common.AppError)
	Delete(ctx context.Context, id uuid.UUID) common.AppError
}

type companyRepository struct {
	db *sql.DB
	l  *slog.Logger
}

// NewCompanyRepository creates a new instance of CompanyRepository.
func NewCompanyRepository(db *sql.DB, logger *slog.Logger) CompanyRepository {
	return &companyRepository{
		db: db,
		l:  logger,
	}
}

// Create inserts a new company record into the database.
// Performs a case-insensitive check for existing company names before insertion.
// Handles potential race conditions by catching unique constraint violations.
// Uses a serializable transaction to ensure data consistency.
func (r *companyRepository) Create(ctx context.Context, company *Company) (*Company, common.AppError) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.l.Error(common.ErrTXBegin, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}
	defer r.rollBackOnError(tx)

	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM companies WHERE LOWER(name) = LOWER($1))", company.Name).Scan(&exists)
	if err != nil {
		r.l.Error("failed to check company existence", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if exists {
		return nil, common.NewConflictError("company with this name already exists")
	}

	query := `
        INSERT INTO companies (id, name, description, amount_of_employees, registered, type)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING created_at, updated_at
    `

	err = tx.QueryRowContext(ctx, query,
		company.ID, company.Name, company.Description, company.AmountOfEmployees,
		company.Registered, company.Type).Scan(&company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, common.NewConflictError("company with this name already exists")
		}

		r.l.Error("failed to create company", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		r.l.Error(common.ErrTxCommit, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return company, nil
}

// FindByID retrieves a company by its UUID.
// Returns NotFoundError if the company doesn't exist.
// Returns a not found error if the company is found but has been deleted.
func (r *companyRepository) FindByID(ctx context.Context, id uuid.UUID) (*Company, common.AppError) {
	query := `
    SELECT id, name, description, amount_of_employees, registered, type, created_at, updated_at, deleted_at
    FROM companies
    WHERE id = $1
    `

	var company Company
	var description sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&company.ID, &company.Name, &description, &company.AmountOfEmployees,
		&company.Registered, &company.Type, &company.CreatedAt, &company.UpdatedAt, &deletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.NewNotFoundError("company not found")
		}
		r.l.Error("failed to get company", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if deletedAt.Valid {
		return nil, common.NewNotFoundError("company has been deleted, please contact support")
	}

	if description.Valid {
		company.Description = &description.String
	}

	return &company, nil
}

// Update modifies an existing company record.
// Uses a serializable transaction to ensure data consistency.
func (r *companyRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*Company, common.AppError) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.l.Error(common.ErrTXBegin, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}
	defer r.rollBackOnError(tx)

	setClause, args := buildUpdateQuery(updates)
	args = append(args, id)

	query := fmt.Sprintf(`
        UPDATE companies
        SET %s
        WHERE id = $%d
        RETURNING id, name, description, amount_of_employees, registered, type, created_at, updated_at
    `, setClause, len(args))

	var company Company
	var description sql.NullString

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&company.ID, &company.Name, &description, &company.AmountOfEmployees,
		&company.Registered, &company.Type, &company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.NewNotFoundError("company not found")
		}
		r.l.Error("failed to update company", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if description.Valid {
		company.Description = &description.String
	}

	if err = tx.Commit(); err != nil {
		r.l.Error(common.ErrTxCommit, "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return &company, nil
}

// Delete performs a soft delete on a company record by setting its deleted_at timestamp.
// Returns NotFoundError if the company doesn't exist or is already deleted.
func (r *companyRepository) Delete(ctx context.Context, id uuid.UUID) common.AppError {
	query := `UPDATE companies SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.l.Error("failed to delete company", "err", err)
		return common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.l.Error("failed to get rows affected", "err", err)
		return common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if rowsAffected == 0 {
		return common.NewNotFoundError("company not found or already deleted")
	}

	return nil
}

// buildUpdateQuery constructs the SET clause and arguments for an UPDATE query.
// Example:
//
//	updates := map[string]any{"name": "New Corp", "registered": true}
//	setClause, args := buildUpdateQuery(updates)
//	// setClause: "name = $1, registered = $2"
//	// args: []any{"New Corp", true}
func buildUpdateQuery(updates map[string]any) (string, []any) {
	setClauses := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates))
	i := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i))
		args = append(args, value)
		i++
	}

	return strings.Join(setClauses, ", "), args
}

// rollBackOnError attempts to roll back a transaction if an error occurred.
// Logs any rollback errors for debugging purposes.
func (r *companyRepository) rollBackOnError(tx *sql.Tx) {
	if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
		r.l.Error(common.ErrTXRollback, "rbErr", rbErr)
	}
}
