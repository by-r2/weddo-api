package wedding

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	weddingRepo   repository.WeddingRepository
	userRepo      repository.UserRepository
	jwtSecret     string
	jwtExpH       int
	ensureCashTpl func(context.Context, string) error
}

func NewUseCase(wr repository.WeddingRepository, ur repository.UserRepository, jwtSecret string, jwtExpH int, ensureCashTpl func(context.Context, string) error) *UseCase {
	return &UseCase{
		weddingRepo: wr, userRepo: ur, jwtSecret: jwtSecret, jwtExpH: jwtExpH,
		ensureCashTpl: ensureCashTpl,
	}
}

// AuthResult agrupa os dados retornados após autenticação ou registro.
type AuthResult struct {
	Token   string
	Wedding *entity.Wedding
	User    *entity.User
}

// Authenticate valida email+senha e retorna JWT.
func (uc *UseCase) Authenticate(ctx context.Context, email, password string) (*AuthResult, error) {
	u, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, entity.ErrUnauthorized
	}

	if u.PasswordHash == "" {
		return nil, entity.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, entity.ErrUnauthorized
	}

	if u.WeddingID == "" {
		return nil, entity.ErrUnauthorized
	}

	w, err := uc.weddingRepo.FindByID(ctx, u.WeddingID)
	if err != nil || !w.Active {
		return nil, entity.ErrUnauthorized
	}

	token, err := uc.signToken(u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, Wedding: w, User: u}, nil
}

// AuthenticateGoogle valida um Google ID Token já decodificado e retorna JWT.
// Se o user existe mas não tem wedding (convite pendente com Google), retorna erro.
func (uc *UseCase) AuthenticateGoogle(ctx context.Context, googleID, email, name, picture string) (*AuthResult, error) {
	u, err := uc.userRepo.FindByGoogleID(ctx, googleID)
	if err != nil {
		u, err = uc.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, entity.ErrUnauthorized
		}
		u.GoogleID = googleID
		u.Name = name
		u.AvatarURL = picture
		u.UpdatedAt = time.Now()
		if err := uc.userRepo.Update(ctx, u); err != nil {
			return nil, fmt.Errorf("wedding.AuthenticateGoogle: link google: %w", err)
		}
	}

	if u.WeddingID == "" {
		return nil, entity.ErrUnauthorized
	}

	w, err := uc.weddingRepo.FindByID(ctx, u.WeddingID)
	if err != nil || !w.Active {
		return nil, entity.ErrUnauthorized
	}

	token, err := uc.signToken(u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, Wedding: w, User: u}, nil
}

type RegisterInput struct {
	Partner1Name string
	Partner2Name string
	Email        string
	Password     string
	Date         string
	Slug         string
}

// Register cria wedding + user com email/senha e retorna JWT.
func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	if _, err := uc.userRepo.FindByEmail(ctx, input.Email); err == nil {
		return nil, entity.ErrAlreadyExists
	}

	w, err := uc.createWedding(ctx, input.Partner1Name, input.Partner2Name, input.Date, input.Slug)
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("wedding.Register: hash password: %w", err)
	}

	now := time.Now()
	u := &entity.User{
		ID:           uuid.New().String(),
		WeddingID:    w.ID,
		Name:         input.Partner1Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("wedding.Register: create user: %w", err)
	}

	token, err := uc.signToken(u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, Wedding: w, User: u}, nil
}

type RegisterGoogleInput struct {
	Partner1Name string
	Partner2Name string
	Date         string
	Slug         string
	GoogleID     string
	Email        string
	Name         string
	Picture      string
}

// RegisterGoogle cria wedding + user com conta Google e retorna JWT.
func (uc *UseCase) RegisterGoogle(ctx context.Context, input RegisterGoogleInput) (*AuthResult, error) {
	if _, err := uc.userRepo.FindByEmail(ctx, input.Email); err == nil {
		return nil, entity.ErrAlreadyExists
	}
	if _, err := uc.userRepo.FindByGoogleID(ctx, input.GoogleID); err == nil {
		return nil, entity.ErrAlreadyExists
	}

	w, err := uc.createWedding(ctx, input.Partner1Name, input.Partner2Name, input.Date, input.Slug)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	u := &entity.User{
		ID:        uuid.New().String(),
		WeddingID: w.ID,
		Name:      input.Name,
		Email:     input.Email,
		AvatarURL: input.Picture,
		GoogleID:  input.GoogleID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("wedding.RegisterGoogle: create user: %w", err)
	}

	token, err := uc.signToken(u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, Wedding: w, User: u}, nil
}

// Seed cria wedding + user se não existir (usado no boot para o primeiro tenant).
func (uc *UseCase) Seed(ctx context.Context, slug, title, date, partner1, partner2, email, password string) error {
	if _, err := uc.userRepo.FindByEmail(ctx, email); err == nil {
		return nil
	}

	now := time.Now()
	w := &entity.Wedding{
		ID:           uuid.New().String(),
		Slug:         slug,
		Title:        title,
		Date:         date,
		Partner1Name: partner1,
		Partner2Name: partner2,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.weddingRepo.Create(ctx, w); err != nil {
		return fmt.Errorf("wedding.Seed: create wedding: %w", err)
	}

	if uc.ensureCashTpl != nil {
		if err := uc.ensureCashTpl(ctx, w.ID); err != nil {
			return fmt.Errorf("wedding.Seed: cash gift template: %w", err)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("wedding.Seed: hash password: %w", err)
	}

	u := &entity.User{
		ID:           uuid.New().String(),
		WeddingID:    w.ID,
		Name:         partner1,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		return fmt.Errorf("wedding.Seed: create user: %w", err)
	}

	return nil
}

func (uc *UseCase) createWedding(ctx context.Context, partner1, partner2, date, slug string) (*entity.Wedding, error) {
	if slug == "" {
		slug = generateSlug(partner1, partner2)
	}

	if _, err := uc.weddingRepo.FindBySlug(ctx, slug); err == nil {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	title := fmt.Sprintf("Casamento %s & %s", partner1, partner2)
	now := time.Now()

	w := &entity.Wedding{
		ID:           uuid.New().String(),
		Slug:         slug,
		Title:        title,
		Date:         date,
		Partner1Name: partner1,
		Partner2Name: partner2,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.weddingRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("wedding: create wedding: %w", err)
	}

	if uc.ensureCashTpl != nil {
		if err := uc.ensureCashTpl(ctx, w.ID); err != nil {
			return nil, fmt.Errorf("wedding: cash gift template: %w", err)
		}
	}

	return w, nil
}

func (uc *UseCase) signToken(u *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"wedding_id": u.WeddingID,
		"user_id":    u.ID,
		"email":      u.Email,
		"exp":        time.Now().Add(time.Duration(uc.jwtExpH) * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("wedding.signToken: %w", err)
	}
	return signed, nil
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(partner1, partner2 string) string {
	raw := strings.ToLower(partner1) + "-e-" + strings.ToLower(partner2)
	slug := nonAlphanumeric.ReplaceAllString(raw, "-")
	return strings.Trim(slug, "-")
}
