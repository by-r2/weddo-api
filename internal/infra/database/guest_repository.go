package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
)

type guestRepository struct {
	db *sql.DB
}

func NewGuestRepository(db *sql.DB) repository.GuestRepository {
	return &guestRepository{db: db}
}

func (r *guestRepository) Create(ctx context.Context, g *entity.Guest) error {
	query := `
		INSERT INTO guests (id, invitation_id, wedding_id, name, phone, email, status, confirmed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		g.ID, g.InvitationID, g.WeddingID, g.Name, g.Phone, g.Email,
		g.Status, g.ConfirmedAt, g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("guestRepository.Create: %w", err)
	}
	return nil
}

func (r *guestRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Guest, error) {
	query := `
		SELECT id, invitation_id, wedding_id, name, phone, email, status, confirmed_at, created_at, updated_at
		FROM guests WHERE wedding_id = ? AND id = ?`

	return r.scanGuest(r.db.QueryRowContext(ctx, query, weddingID, id))
}

func (r *guestRepository) FindByName(ctx context.Context, weddingID, name string) (*entity.Guest, error) {
	query := `
		SELECT id, invitation_id, wedding_id, name, phone, email, status, confirmed_at, created_at, updated_at
		FROM guests WHERE wedding_id = ? AND name = ? COLLATE NOCASE`

	return r.scanGuest(r.db.QueryRowContext(ctx, query, weddingID, name))
}

func (r *guestRepository) ListByInvitation(ctx context.Context, weddingID, invitationID string) ([]entity.Guest, error) {
	query := `
		SELECT id, invitation_id, wedding_id, name, phone, email, status, confirmed_at, created_at, updated_at
		FROM guests WHERE wedding_id = ? AND invitation_id = ? ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, weddingID, invitationID)
	if err != nil {
		return nil, fmt.Errorf("guestRepository.ListByInvitation: %w", err)
	}
	defer rows.Close()

	return r.scanGuests(rows)
}

func (r *guestRepository) List(ctx context.Context, weddingID string, page, perPage int, status, search string) ([]entity.Guest, int, error) {
	countQuery := `SELECT COUNT(*) FROM guests WHERE wedding_id = ?`
	listQuery := `
		SELECT id, invitation_id, wedding_id, name, phone, email, status, confirmed_at, created_at, updated_at
		FROM guests WHERE wedding_id = ?`

	args := []any{weddingID}

	if status != "" {
		filter := ` AND status = ?`
		countQuery += filter
		listQuery += filter
		args = append(args, status)
	}
	if search != "" {
		filter := ` AND name LIKE ?`
		countQuery += filter
		listQuery += filter
		args = append(args, "%"+search+"%")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("guestRepository.List: count: %w", err)
	}

	listQuery += ` ORDER BY name LIMIT ? OFFSET ?`
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("guestRepository.List: query: %w", err)
	}
	defer rows.Close()

	guests, err := r.scanGuests(rows)
	if err != nil {
		return nil, 0, err
	}
	return guests, total, nil
}

func (r *guestRepository) Update(ctx context.Context, g *entity.Guest) error {
	query := `
		UPDATE guests SET name = ?, phone = ?, email = ?, status = ?, confirmed_at = ?, updated_at = ?
		WHERE wedding_id = ? AND id = ?`

	res, err := r.db.ExecContext(ctx, query,
		g.Name, g.Phone, g.Email, g.Status, g.ConfirmedAt, g.UpdatedAt,
		g.WeddingID, g.ID,
	)
	if err != nil {
		return fmt.Errorf("guestRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *guestRepository) Delete(ctx context.Context, weddingID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM guests WHERE wedding_id = ? AND id = ?`, weddingID, id)
	if err != nil {
		return fmt.Errorf("guestRepository.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *guestRepository) CountByWedding(ctx context.Context, weddingID string) (total, confirmed, pending, declined int, err error) {
	query := `
		SELECT
			COUNT(*),
			COUNT(CASE WHEN status = 'confirmed' THEN 1 END),
			COUNT(CASE WHEN status = 'pending' THEN 1 END),
			COUNT(CASE WHEN status = 'declined' THEN 1 END)
		FROM guests WHERE wedding_id = ?`

	err = r.db.QueryRowContext(ctx, query, weddingID).Scan(&total, &confirmed, &pending, &declined)
	if err != nil {
		err = fmt.Errorf("guestRepository.CountByWedding: %w", err)
	}
	return
}

func (r *guestRepository) scanGuest(row *sql.Row) (*entity.Guest, error) {
	var g entity.Guest
	err := row.Scan(
		&g.ID, &g.InvitationID, &g.WeddingID, &g.Name, &g.Phone, &g.Email,
		&g.Status, &g.ConfirmedAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("guestRepository.scanGuest: %w", err)
	}
	return &g, nil
}

func (r *guestRepository) scanGuests(rows *sql.Rows) ([]entity.Guest, error) {
	var guests []entity.Guest
	for rows.Next() {
		var g entity.Guest
		if err := rows.Scan(
			&g.ID, &g.InvitationID, &g.WeddingID, &g.Name, &g.Phone, &g.Email,
			&g.Status, &g.ConfirmedAt, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("guestRepository.scanGuests: %w", err)
		}
		guests = append(guests, g)
	}
	return guests, nil
}
