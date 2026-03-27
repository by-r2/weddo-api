package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type UseCase struct {
	userRepo repository.UserRepository
}

func NewUseCase(ur repository.UserRepository) *UseCase {
	return &UseCase{userRepo: ur}
}

// Invite cria um user pendente vinculado ao wedding.
// Se o email já existir com wedding_id preenchido, retorna ErrAlreadyExists.
// Se existir sem wedding (cenário improvável), vincula ao wedding.
func (uc *UseCase) Invite(ctx context.Context, weddingID, email, name string) (*entity.User, error) {
	existing, err := uc.userRepo.FindByEmail(ctx, email)
	if err == nil {
		if existing.WeddingID != "" {
			return nil, entity.ErrAlreadyExists
		}
		existing.WeddingID = weddingID
		if name != "" {
			existing.Name = name
		}
		existing.UpdatedAt = time.Now()
		if err := uc.userRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("user.Invite: link existing: %w", err)
		}
		return existing, nil
	}

	now := time.Now()
	u := &entity.User{
		ID:        uuid.New().String(),
		WeddingID: weddingID,
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("user.Invite: %w", err)
	}

	return u, nil
}

func (uc *UseCase) List(ctx context.Context, weddingID string) ([]entity.User, error) {
	return uc.userRepo.ListByWedding(ctx, weddingID)
}

func (uc *UseCase) Remove(ctx context.Context, weddingID, userID string) error {
	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.WeddingID != weddingID {
		return entity.ErrNotFound
	}

	users, err := uc.userRepo.ListByWedding(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("user.Remove: list: %w", err)
	}
	activeCount := 0
	for _, usr := range users {
		if usr.PasswordHash != "" || usr.GoogleID != "" {
			activeCount++
		}
	}
	if activeCount <= 1 && (u.PasswordHash != "" || u.GoogleID != "") {
		return fmt.Errorf("user.Remove: %w", entity.ErrUnauthorized)
	}

	return uc.userRepo.Delete(ctx, userID)
}
