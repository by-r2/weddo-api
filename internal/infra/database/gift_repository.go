package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type giftRepository struct {
	db *sql.DB
}

func NewGiftRepository(db *sql.DB) repository.GiftRepository {
	return &giftRepository{db: db}
}

func (r *giftRepository) Create(ctx context.Context, g *entity.Gift) error {
	query := `
		INSERT INTO gifts (id, wedding_id, name, description, price, image_url, category, status, kind, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	kind := g.Kind
	if kind == "" {
		kind = entity.GiftKindCatalog
	}

	_, err := r.db.ExecContext(ctx, query,
		g.ID, g.WeddingID, g.Name, g.Description, g.Price, g.ImageURL,
		g.Category, g.Status, kind, g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("giftRepository.Create: %w", err)
	}
	return nil
}

func (r *giftRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Gift, error) {
	query := `
		SELECT id, wedding_id, name, description, price, image_url, category, status, kind, created_at, updated_at
		FROM gifts WHERE wedding_id = $1 AND id = $2`

	var g entity.Gift
	var kind sql.NullString
	err := r.db.QueryRowContext(ctx, query, weddingID, id).Scan(
		&g.ID, &g.WeddingID, &g.Name, &g.Description, &g.Price, &g.ImageURL,
		&g.Category, &g.Status, &kind, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("giftRepository.FindByID: %w", err)
	}
	if kind.Valid {
		g.Kind = entity.GiftKind(kind.String)
	} else {
		g.Kind = entity.GiftKindCatalog
	}
	return &g, nil
}

func (r *giftRepository) FindCashTemplateByWeddingID(ctx context.Context, weddingID string) (*entity.Gift, error) {
	query := `
		SELECT id, wedding_id, name, description, price, image_url, category, status, kind, created_at, updated_at
		FROM gifts WHERE wedding_id = $1 AND kind = 'cash_template' LIMIT 1`
	var g entity.Gift
	var kind sql.NullString
	err := r.db.QueryRowContext(ctx, query, weddingID).Scan(
		&g.ID, &g.WeddingID, &g.Name, &g.Description, &g.Price, &g.ImageURL,
		&g.Category, &g.Status, &kind, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("giftRepository.FindCashTemplateByWeddingID: %w", err)
	}
	if kind.Valid {
		g.Kind = entity.GiftKind(kind.String)
	}
	return &g, nil
}

func (r *giftRepository) List(ctx context.Context, weddingID string, page, perPage int, category, status, search string, catalogOnly bool) ([]entity.Gift, int, error) {
	countQuery := `SELECT COUNT(*) FROM gifts WHERE wedding_id = $1`
	listQuery := `
		SELECT id, wedding_id, name, description, price, image_url, category, status, kind, created_at, updated_at
		FROM gifts WHERE wedding_id = $1`

	args := []any{weddingID}
	paramIdx := 2

	if catalogOnly {
		f := fmt.Sprintf(` AND kind = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, entity.GiftKindCatalog)
		paramIdx++
	}
	if category != "" {
		f := fmt.Sprintf(` AND category = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, category)
		paramIdx++
	}
	if status != "" {
		f := fmt.Sprintf(` AND status = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, status)
		paramIdx++
	}
	if search != "" {
		f := fmt.Sprintf(` AND name ILIKE $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, "%"+search+"%")
		paramIdx++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("giftRepository.List: count: %w", err)
	}

	listQuery += fmt.Sprintf(` ORDER BY category, name LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("giftRepository.List: query: %w", err)
	}
	defer rows.Close()

	var gifts []entity.Gift
	for rows.Next() {
		var g entity.Gift
		var kind sql.NullString
		if err := rows.Scan(
			&g.ID, &g.WeddingID, &g.Name, &g.Description, &g.Price, &g.ImageURL,
			&g.Category, &g.Status, &kind, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("giftRepository.List: scan: %w", err)
		}
		if kind.Valid {
			g.Kind = entity.GiftKind(kind.String)
		} else {
			g.Kind = entity.GiftKindCatalog
		}
		gifts = append(gifts, g)
	}
	return gifts, total, nil
}

func (r *giftRepository) Update(ctx context.Context, g *entity.Gift) error {
	query := `
		UPDATE gifts SET name = $1, description = $2, price = $3, image_url = $4, category = $5, status = $6, updated_at = $7
		WHERE wedding_id = $8 AND id = $9`

	res, err := r.db.ExecContext(ctx, query,
		g.Name, g.Description, g.Price, g.ImageURL, g.Category, g.Status, g.UpdatedAt,
		g.WeddingID, g.ID,
	)
	if err != nil {
		return fmt.Errorf("giftRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *giftRepository) Delete(ctx context.Context, weddingID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM gifts WHERE wedding_id = $1 AND id = $2`, weddingID, id)
	if err != nil {
		return fmt.Errorf("giftRepository.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *giftRepository) ListCategories(ctx context.Context, weddingID string) ([]string, error) {
	const q = `
		SELECT DISTINCT TRIM(BOTH FROM category) AS cat
		FROM gifts
		WHERE wedding_id = $1 AND kind = $2
		  AND LENGTH(TRIM(BOTH FROM category)) > 0
		ORDER BY LOWER(TRIM(BOTH FROM category))`

	rows, err := r.db.QueryContext(ctx, q, weddingID, entity.GiftKindCatalog)
	if err != nil {
		return nil, fmt.Errorf("giftRepository.ListCategories: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, fmt.Errorf("giftRepository.ListCategories: scan: %w", err)
		}
		out = append(out, cat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("giftRepository.ListCategories: rows: %w", err)
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func (r *giftRepository) CountByWedding(ctx context.Context, weddingID string) (total, available, purchased int, err error) {
	query := `
		SELECT
			COUNT(*),
			COUNT(CASE WHEN status = 'available' THEN 1 END),
			COUNT(CASE WHEN status = 'purchased' THEN 1 END)
		FROM gifts WHERE wedding_id = $1 AND kind = $2`

	err = r.db.QueryRowContext(ctx, query, weddingID, entity.GiftKindCatalog).Scan(&total, &available, &purchased)
	if err != nil {
		err = fmt.Errorf("giftRepository.CountByWedding: %w", err)
	}
	return
}
