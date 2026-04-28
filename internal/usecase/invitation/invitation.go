package invitation

import (
	"context"
	"fmt"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"github.com/google/uuid"
)

type UseCase struct {
	invRepo   repository.InvitationRepository
	guestRepo repository.GuestRepository
	txManager repository.TransactionManager
}

func NewUseCase(ir repository.InvitationRepository, gr repository.GuestRepository, txManager repository.TransactionManager) *UseCase {
	return &UseCase{invRepo: ir, guestRepo: gr, txManager: txManager}
}

type CreateGuestInput struct {
	Name   string
	Phone  string
	Email  string
	Status string
}

type CreateInput struct {
	WeddingID string
	Code      string
	Label     string
	MaxGuests int
	Notes     string
	Guests    []CreateGuestInput
}

func (uc *UseCase) Create(ctx context.Context, input CreateInput) (*entity.Invitation, error) {
	now := time.Now()
	inv := &entity.Invitation{
		ID:        uuid.New().String(),
		WeddingID: input.WeddingID,
		Code:      input.Code,
		Label:     input.Label,
		MaxGuests: input.MaxGuests,
		Notes:     input.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.invRepo.Create(txCtx, inv); err != nil {
			return fmt.Errorf("invitation.Create: %w", err)
		}

		for _, gi := range input.Guests {
			status := sanitizeGuestStatus(gi.Status)
			guest := &entity.Guest{
				ID:           uuid.New().String(),
				InvitationID: inv.ID,
				WeddingID:    input.WeddingID,
				Name:         gi.Name,
				Phone:        gi.Phone,
				Email:        gi.Email,
				Status:       status,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if status == entity.GuestStatusConfirmed {
				confirmedAt := now
				guest.ConfirmedAt = &confirmedAt
			}
			if err := uc.guestRepo.Create(txCtx, guest); err != nil {
				return fmt.Errorf("invitation.Create: create guest %q: %w", gi.Name, err)
			}
			inv.Guests = append(inv.Guests, *guest)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (uc *UseCase) FindByID(ctx context.Context, weddingID, id string) (*entity.Invitation, error) {
	inv, err := uc.invRepo.FindByID(ctx, weddingID, id)
	if err != nil {
		return nil, err
	}

	guests, err := uc.guestRepo.ListByInvitation(ctx, weddingID, id)
	if err != nil {
		return nil, fmt.Errorf("invitation.FindByID: list guests: %w", err)
	}
	inv.Guests = guests

	return inv, nil
}

func (uc *UseCase) List(ctx context.Context, weddingID string, page, perPage int, search string) ([]entity.Invitation, int, error) {
	return uc.invRepo.List(ctx, weddingID, page, perPage, search)
}

type UpdateInput struct {
	WeddingID string
	ID        string
	Code      string
	Label     string
	MaxGuests int
	Notes     string
}

func (uc *UseCase) Update(ctx context.Context, input UpdateInput) (*entity.Invitation, error) {
	inv, err := uc.invRepo.FindByID(ctx, input.WeddingID, input.ID)
	if err != nil {
		return nil, err
	}

	inv.Code = input.Code
	inv.Label = input.Label
	inv.MaxGuests = input.MaxGuests
	inv.Notes = input.Notes
	inv.UpdatedAt = time.Now()

	if err := uc.invRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("invitation.Update: %w", err)
	}
	return inv, nil
}

func (uc *UseCase) Delete(ctx context.Context, weddingID, id string) error {
	return uc.invRepo.Delete(ctx, weddingID, id)
}

// AddGuest adiciona um convidado a um convite existente.
func (uc *UseCase) AddGuest(ctx context.Context, weddingID, invitationID, name, phone, email, statusRaw string) (*entity.Guest, error) {
	now := time.Now()
	status := sanitizeGuestStatus(statusRaw)
	guest := &entity.Guest{
		ID:           uuid.New().String(),
		InvitationID: invitationID,
		WeddingID:    weddingID,
		Name:         name,
		Phone:        phone,
		Email:        email,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if status == entity.GuestStatusConfirmed {
		confirmedAt := now
		guest.ConfirmedAt = &confirmedAt
	}

	err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		inv, err := uc.invRepo.FindByID(txCtx, weddingID, invitationID)
		if err != nil {
			return err
		}

		if err := uc.guestRepo.Create(txCtx, guest); err != nil {
			return fmt.Errorf("invitation.AddGuest: %w", err)
		}

		inv.UpdatedAt = now
		if err := uc.invRepo.Update(txCtx, inv); err != nil {
			return fmt.Errorf("invitation.AddGuest: touch invitation: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return guest, nil
}

func sanitizeGuestStatus(raw string) entity.GuestStatus {
	status := entity.GuestStatus(raw)
	switch status {
	case entity.GuestStatusConfirmed, entity.GuestStatusDeclined:
		return status
	default:
		return entity.GuestStatusPending
	}
}
