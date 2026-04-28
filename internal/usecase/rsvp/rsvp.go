package rsvp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type UseCase struct {
	guestRepo      repository.GuestRepository
	invitationRepo repository.InvitationRepository
}

func NewUseCase(gr repository.GuestRepository, ir repository.InvitationRepository) *UseCase {
	return &UseCase{guestRepo: gr, invitationRepo: ir}
}

// Confirm registra a confirmação de presença de um convidado no convite identificado por code.
func (uc *UseCase) Confirm(ctx context.Context, weddingID, invitationCode, guestName string) (*entity.Guest, *entity.Invitation, bool, error) {
	inv, err := uc.invitationRepo.FindByCode(ctx, weddingID, invitationCode)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, false, ErrInvitationNotFound
		}
		return nil, nil, false, err
	}

	guest, err := uc.guestRepo.FindByNameInInvitation(ctx, weddingID, inv.ID, guestName)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, false, ErrGuestNotFoundOnInvitation
		}
		return nil, nil, false, err
	}

	alreadyConfirmed := guest.Status == entity.GuestStatusConfirmed

	if !alreadyConfirmed {
		now := time.Now()
		guest.Status = entity.GuestStatusConfirmed
		guest.ConfirmedAt = &now
		guest.UpdatedAt = now

		if err := uc.guestRepo.Update(ctx, guest); err != nil {
			return nil, nil, false, fmt.Errorf("rsvp.Confirm: %w", err)
		}
	}

	return guest, inv, alreadyConfirmed, nil
}

// FindInvitationByCode busca o convite pelo code (único por wedding) e lista todos os guests.
func (uc *UseCase) FindInvitationByCode(ctx context.Context, weddingID, code string) (*entity.Invitation, []entity.Guest, error) {
	inv, err := uc.invitationRepo.FindByCode(ctx, weddingID, code)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, ErrInvitationNotFound
		}
		return nil, nil, err
	}

	guests, err := uc.guestRepo.ListByInvitation(ctx, weddingID, inv.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("rsvp.FindInvitationByCode: list guests: %w", err)
	}

	return inv, guests, nil
}
