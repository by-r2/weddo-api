package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreateWithItems(ctx context.Context, p *entity.Payment, items []entity.PaymentItem) error {
	if len(items) == 0 {
		return fmt.Errorf("paymentRepository.CreateWithItems: no items")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("paymentRepository.CreateWithItems: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qpay := `
		INSERT INTO payments (id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	if _, err := tx.ExecContext(ctx, qpay,
		p.ID, p.WeddingID, p.ProviderID, p.Amount, p.Status, p.PaymentMethod,
		p.PayerName, p.PayerEmail, p.Message, p.PixQRCode, p.PixExpiration,
		p.PaidAt, p.CreatedAt, p.UpdatedAt,
	); err != nil {
		return fmt.Errorf("paymentRepository.CreateWithItems: insert payment: %w", err)
	}

	qitem := `
		INSERT INTO payment_items (id, payment_id, gift_id, amount, custom_name, custom_description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for _, it := range items {
		var cn, cd any
		if strings.TrimSpace(it.CustomName) != "" {
			cn = strings.TrimSpace(it.CustomName)
		}
		if strings.TrimSpace(it.CustomDescription) != "" {
			cd = strings.TrimSpace(it.CustomDescription)
		}
		if _, err := tx.ExecContext(ctx, qitem,
			it.ID, it.PaymentID, it.GiftID, it.Amount, cn, cd, it.CreatedAt,
		); err != nil {
			return fmt.Errorf("paymentRepository.CreateWithItems: insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("paymentRepository.CreateWithItems: commit: %w", err)
	}
	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error) {
	query := `
		SELECT id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE wedding_id = $1 AND id = $2`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, weddingID, id))
}

func (r *paymentRepository) FindByIDAny(ctx context.Context, id string) (*entity.Payment, error) {
	query := `
		SELECT id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE id = $1`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, id))
}

func (r *paymentRepository) FindByProviderID(ctx context.Context, providerID string) (*entity.Payment, error) {
	query := `
		SELECT id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE provider_id = $1`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, providerID))
}

func (r *paymentRepository) FindItemsByPaymentID(ctx context.Context, weddingID, paymentID string) ([]entity.PaymentItem, error) {
	query := `
		SELECT pi.id, pi.payment_id, pi.gift_id, pi.amount,
			COALESCE(pi.custom_name,''), COALESCE(pi.custom_description,''), pi.created_at
		FROM payment_items pi
		INNER JOIN payments p ON p.id = pi.payment_id
		WHERE p.wedding_id = $1 AND p.id = $2
		ORDER BY pi.created_at`

	rows, err := r.db.QueryContext(ctx, query, weddingID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("paymentRepository.FindItemsByPaymentID: %w", err)
	}
	defer rows.Close()

	var out []entity.PaymentItem
	for rows.Next() {
		var it entity.PaymentItem
		var cn, cd sql.NullString
		if err := rows.Scan(&it.ID, &it.PaymentID, &it.GiftID, &it.Amount, &cn, &cd, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("paymentRepository.FindItemsByPaymentID: scan: %w", err)
		}
		if cn.Valid {
			it.CustomName = cn.String
		}
		if cd.Valid {
			it.CustomDescription = cd.String
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *paymentRepository) List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error) {
	countQuery := `SELECT COUNT(*) FROM payments WHERE wedding_id = $1`
	listQuery := `
		SELECT id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE wedding_id = $1`

	args := []any{weddingID}
	paramIdx := 2

	if status != "" {
		f := fmt.Sprintf(` AND status = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, status)
		paramIdx++
	}
	if giftID == "cash" {
		existsCash := `
			AND EXISTS (
				SELECT 1 FROM payment_items pi
				INNER JOIN gifts g ON g.id = pi.gift_id AND g.kind = 'cash_template'
				WHERE pi.payment_id = payments.id
			)`
		countQuery += existsCash
		listQuery += existsCash
	} else if giftID != "" {
		existsGift := fmt.Sprintf(`
			AND EXISTS (SELECT 1 FROM payment_items pi WHERE pi.payment_id = payments.id AND pi.gift_id = $%d)`, paramIdx)
		countQuery += existsGift
		listQuery += existsGift
		args = append(args, giftID)
		paramIdx++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("paymentRepository.List: count: %w", err)
	}

	listQuery += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("paymentRepository.List: query: %w", err)
	}
	defer rows.Close()

	var payments []entity.Payment
	for rows.Next() {
		p, err := scanPaymentFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("paymentRepository.List: scan: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *paymentRepository) Update(ctx context.Context, p *entity.Payment) error {
	query := `
		UPDATE payments SET provider_id = $1, status = $2, pix_qr_code = $3, pix_expiration = $4,
			paid_at = $5, updated_at = $6
		WHERE id = $7`

	res, err := r.db.ExecContext(ctx, query,
		p.ProviderID, p.Status, p.PixQRCode, p.PixExpiration, p.PaidAt, p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("paymentRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *paymentRepository) SumByWedding(ctx context.Context, weddingID string) (totalRevenue float64, totalPayments int, err error) {
	query := `
		SELECT COALESCE(SUM(amount), 0), COUNT(*)
		FROM payments WHERE wedding_id = $1 AND status = 'approved'`

	err = r.db.QueryRowContext(ctx, query, weddingID).Scan(&totalRevenue, &totalPayments)
	if err != nil {
		err = fmt.Errorf("paymentRepository.SumByWedding: %w", err)
	}
	return
}

func (r *paymentRepository) scanPayment(row *sql.Row) (*entity.Payment, error) {
	p, err := scanPaymentFields(row.Scan)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("paymentRepository.scanPayment: %w", err)
	}
	return &p, nil
}

func scanPaymentFromRows(rows *sql.Rows) (entity.Payment, error) {
	return scanPaymentFields(rows.Scan)
}

func scanPaymentFields(scan func(dest ...any) error) (entity.Payment, error) {
	var p entity.Payment
	err := scan(
		&p.ID, &p.WeddingID, &p.ProviderID, &p.Amount, &p.Status, &p.PaymentMethod,
		&p.PayerName, &p.PayerEmail, &p.Message, &p.PixQRCode, &p.PixExpiration,
		&p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}
