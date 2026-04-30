package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func giftListOrderBy(sortBy, sortDir string) string {
	switch sortBy {
	case "price":
		if sortDir == "desc" {
			return "price DESC, name ASC"
		}
		return "price ASC, name ASC"
	case "name":
		if sortDir == "desc" {
			return "name DESC"
		}
		return "name ASC"
	default:
		return "category ASC, name ASC"
	}
}

func (r *giftRepository) List(ctx context.Context, weddingID string, p repository.GiftListParams) ([]entity.Gift, int, error) {
	countQuery := `SELECT COUNT(*) FROM gifts WHERE wedding_id = $1`
	listQuery := `
		SELECT id, wedding_id, name, description, price, image_url, category, status, kind, created_at, updated_at
		FROM gifts WHERE wedding_id = $1`

	args := []any{weddingID}
	paramIdx := 2

	page := p.Page
	perPage := p.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	if p.CatalogOnly {
		f := fmt.Sprintf(` AND kind = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, entity.GiftKindCatalog)
		paramIdx++
	}
	if len(p.Categories) > 0 {
		placeholders := make([]string, len(p.Categories))
		for i := range p.Categories {
			placeholders[i] = fmt.Sprintf("$%d", paramIdx+i)
		}
		clause := fmt.Sprintf(` AND category IN (%s)`, strings.Join(placeholders, ","))
		countQuery += clause
		listQuery += clause
		for _, c := range p.Categories {
			args = append(args, c)
		}
		paramIdx += len(p.Categories)
	}
	if p.Status != "" {
		f := fmt.Sprintf(` AND status = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, p.Status)
		paramIdx++
	}
	if p.Search != "" {
		f := fmt.Sprintf(` AND name ILIKE $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, "%"+p.Search+"%")
		paramIdx++
	}
	if p.MinPrice != nil {
		f := fmt.Sprintf(` AND price >= $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, *p.MinPrice)
		paramIdx++
	}
	if p.MaxPrice != nil {
		f := fmt.Sprintf(` AND price <= $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, *p.MaxPrice)
		paramIdx++
	}

	sortBy := strings.TrimSpace(strings.ToLower(p.SortBy))
	if sortBy == "" {
		sortBy = "recommended"
	}
	sortDir := strings.TrimSpace(strings.ToLower(p.SortDir))
	if sortDir == "" {
		sortDir = "asc"
	}
	if sortBy != "price" && sortBy != "name" {
		sortBy = "recommended"
	}
	order := giftListOrderBy(sortBy, sortDir)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("giftRepository.List: count: %w", err)
	}

	listQuery += fmt.Sprintf(` ORDER BY %s LIMIT $%d OFFSET $%d`, order, paramIdx, paramIdx+1)
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

func (r *giftRepository) ListCategories(ctx context.Context, weddingID string) ([]repository.GiftCategoryCount, error) {
	const q = `
		SELECT TRIM(BOTH FROM category) AS cat, COUNT(*) AS total
		FROM gifts
		WHERE wedding_id = $1
		  AND kind = $2
		  AND LENGTH(TRIM(BOTH FROM category)) > 0
		GROUP BY TRIM(BOTH FROM category)
		ORDER BY LOWER(TRIM(BOTH FROM category))`

	rows, err := r.db.QueryContext(ctx, q, weddingID, entity.GiftKindCatalog)
	if err != nil {
		return nil, fmt.Errorf("giftRepository.ListCategories: %w", err)
	}
	defer rows.Close()

	var out []repository.GiftCategoryCount
	for rows.Next() {
		var item repository.GiftCategoryCount
		if err := rows.Scan(&item.Category, &item.Count); err != nil {
			return nil, fmt.Errorf("giftRepository.ListCategories: scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("giftRepository.ListCategories: rows: %w", err)
	}
	if out == nil {
		out = []repository.GiftCategoryCount{}
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
