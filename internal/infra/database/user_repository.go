package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *entity.User) error {
	query := `
		INSERT INTO users (id, wedding_id, name, email, avatar_url, google_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		u.ID, nullIfEmpty(u.WeddingID), u.Name, u.Email, u.AvatarURL,
		nullIfEmpty(u.GoogleID), nullIfEmpty(u.PasswordHash),
		u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("userRepository.Create: %w", err)
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return r.findBy(ctx, "id", id)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.findBy(ctx, "email", email)
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	return r.findBy(ctx, "google_id", googleID)
}

func (r *userRepository) ListByWedding(ctx context.Context, weddingID string) ([]entity.User, error) {
	query := `
		SELECT id, wedding_id, name, email, avatar_url, google_id, password_hash, created_at, updated_at
		FROM users WHERE wedding_id = $1 ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, weddingID)
	if err != nil {
		return nil, fmt.Errorf("userRepository.ListByWedding: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("userRepository.ListByWedding: scan: %w", err)
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *userRepository) Update(ctx context.Context, u *entity.User) error {
	query := `
		UPDATE users
		SET wedding_id = $2, name = $3, email = $4, avatar_url = $5,
		    google_id = $6, password_hash = $7, updated_at = $8
		WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query,
		u.ID, nullIfEmpty(u.WeddingID), u.Name, u.Email, u.AvatarURL,
		nullIfEmpty(u.GoogleID), nullIfEmpty(u.PasswordHash), u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *userRepository) findBy(ctx context.Context, column, value string) (*entity.User, error) {
	query := fmt.Sprintf(`
		SELECT id, wedding_id, name, email, avatar_url, google_id, password_hash, created_at, updated_at
		FROM users WHERE %s = $1`, column)

	row := r.db.QueryRowContext(ctx, query, value)
	u, err := scanUserRow(row)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepository.findBy(%s): %w", column, err)
	}
	return u, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUserFromRow(s scanner) (*entity.User, error) {
	var u entity.User
	var weddingID, googleID, passwordHash sql.NullString

	err := s.Scan(
		&u.ID, &weddingID, &u.Name, &u.Email, &u.AvatarURL,
		&googleID, &passwordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.WeddingID = weddingID.String
	u.GoogleID = googleID.String
	u.PasswordHash = passwordHash.String
	return &u, nil
}

func scanUser(rows *sql.Rows) (*entity.User, error) {
	return scanUserFromRow(rows)
}

func scanUserRow(row *sql.Row) (*entity.User, error) {
	return scanUserFromRow(row)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
