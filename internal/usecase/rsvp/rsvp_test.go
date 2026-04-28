package rsvp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"github.com/by-r2/weddo-api/internal/usecase/rsvp"
)

func TestConfirm_invitationNotFound(t *testing.T) {
	t.Parallel()
	uc := rsvp.NewUseCase(&guestFake{}, &invFake{findByCodeErr: entity.ErrNotFound})
	_, _, _, err := uc.Confirm(context.Background(), "w1", "missing", "João")
	if !errors.Is(err, rsvp.ErrInvitationNotFound) {
		t.Fatalf("got %v, want ErrInvitationNotFound", err)
	}
}

func TestConfirm_guestNotFound(t *testing.T) {
	t.Parallel()
	inv := &entity.Invitation{ID: "inv-1", WeddingID: "w1", Code: "ABC", Label: "Família"}
	uc := rsvp.NewUseCase(
		&guestFake{findErr: entity.ErrNotFound},
		&invFake{findByCode: inv},
	)
	_, _, _, err := uc.Confirm(context.Background(), "w1", "ABC", "Não existe")
	if !errors.Is(err, rsvp.ErrGuestNotFoundOnInvitation) {
		t.Fatalf("got %v, want ErrGuestNotFoundOnInvitation", err)
	}
}

func TestConfirm_success_pending(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	inv := &entity.Invitation{ID: "inv-1", WeddingID: "w1", Code: "ABC", Label: "Família", MaxGuests: 3}
	g := &entity.Guest{
		ID: "g1", InvitationID: inv.ID, WeddingID: "w1", Name: "João",
		Status: entity.GuestStatusPending, CreatedAt: now, UpdatedAt: now,
	}
	guestR := &guestFake{find: g}
	uc := rsvp.NewUseCase(guestR, &invFake{findByCode: inv})

	outGuest, outInv, already, err := uc.Confirm(context.Background(), "w1", "ABC", "João")
	if err != nil {
		t.Fatal(err)
	}
	if already {
		t.Fatal("expected alreadyConfirmed false")
	}
	if outInv.Code != "ABC" {
		t.Fatalf("invitation: %+v", outInv)
	}
	if outGuest.Status != entity.GuestStatusConfirmed || outGuest.ConfirmedAt == nil {
		t.Fatalf("guest not confirmed: %+v", outGuest)
	}
	if guestR.updateN != 1 {
		t.Fatalf("Update calls = %d", guestR.updateN)
	}
}

func TestConfirm_alreadyConfirmed(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	confirmedAt := now.Add(-time.Hour)
	inv := &entity.Invitation{ID: "inv-1", WeddingID: "w1", Code: "ABC", Label: "Família", MaxGuests: 3}
	g := &entity.Guest{
		ID: "g1", InvitationID: inv.ID, WeddingID: "w1", Name: "João",
		Status: entity.GuestStatusConfirmed, ConfirmedAt: &confirmedAt, CreatedAt: now, UpdatedAt: now,
	}
	guestR := &guestFake{find: g}
	uc := rsvp.NewUseCase(guestR, &invFake{findByCode: inv})

	_, _, already, err := uc.Confirm(context.Background(), "w1", "ABC", "João")
	if err != nil {
		t.Fatal(err)
	}
	if !already {
		t.Fatal("expected alreadyConfirmed true")
	}
	if guestR.updateN != 0 {
		t.Fatalf("Update should not run, got %d calls", guestR.updateN)
	}
}

func TestConfirm_updateError(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	inv := &entity.Invitation{ID: "inv-1", WeddingID: "w1", Code: "ABC", Label: "Família", MaxGuests: 3}
	g := &entity.Guest{
		ID: "g1", InvitationID: inv.ID, WeddingID: "w1", Name: "João",
		Status: entity.GuestStatusPending, CreatedAt: now, UpdatedAt: now,
	}
	dbErr := errors.New("db unavailable")
	uc := rsvp.NewUseCase(
		&guestFake{find: g, updateErr: dbErr},
		&invFake{findByCode: inv},
	)
	_, _, _, err := uc.Confirm(context.Background(), "w1", "ABC", "João")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, rsvp.ErrInvitationNotFound) || errors.Is(err, rsvp.ErrGuestNotFoundOnInvitation) {
		t.Fatalf("unexpected sentinel: %v", err)
	}
}

func TestFindInvitationByCode_notFound(t *testing.T) {
	t.Parallel()
	uc := rsvp.NewUseCase(&guestFake{}, &invFake{findByCodeErr: entity.ErrNotFound})
	_, _, err := uc.FindInvitationByCode(context.Background(), "w1", "x")
	if !errors.Is(err, rsvp.ErrInvitationNotFound) {
		t.Fatalf("got %v, want ErrInvitationNotFound", err)
	}
}

func TestFindInvitationByCode_success(t *testing.T) {
	t.Parallel()
	inv := &entity.Invitation{ID: "inv-1", WeddingID: "w1", Code: "ABC", Label: "Família", MaxGuests: 2}
	guests := []entity.Guest{{ID: "g1", Name: "A", Status: entity.GuestStatusPending}}
	uc := rsvp.NewUseCase(
		&guestFake{listGuests: guests},
		&invFake{findByCode: inv},
	)
	outInv, outGuests, err := uc.FindInvitationByCode(context.Background(), "w1", "ABC")
	if err != nil {
		t.Fatal(err)
	}
	if outInv.ID != inv.ID || len(outGuests) != 1 || outGuests[0].Name != "A" {
		t.Fatalf("inv=%+v guests=%+v", outInv, outGuests)
	}
}

func TestFindInvitationByCode_listError(t *testing.T) {
	t.Parallel()
	inv := &entity.Invitation{ID: "inv-1", WeddingID: "w1", Code: "ABC", Label: "Família", MaxGuests: 2}
	listErr := errors.New("list failed")
	uc := rsvp.NewUseCase(
		&guestFake{listErr: listErr},
		&invFake{findByCode: inv},
	)
	_, _, err := uc.FindInvitationByCode(context.Background(), "w1", "ABC")
	if err == nil || !errors.Is(err, listErr) {
		t.Fatalf("got %v, want list error", err)
	}
}

// --- fakes (implementam interfaces completas; só FindByCode / guests usados nos testes acima)

type invFake struct {
	findByCode    *entity.Invitation
	findByCodeErr error
}

func (f *invFake) Create(context.Context, *entity.Invitation) error { return nil }

func (f *invFake) FindByID(context.Context, string, string) (*entity.Invitation, error) {
	return nil, entity.ErrNotFound
}

func (f *invFake) FindByCode(_ context.Context, _, _ string) (*entity.Invitation, error) {
	if f.findByCodeErr != nil {
		return nil, f.findByCodeErr
	}
	if f.findByCode != nil {
		return f.findByCode, nil
	}
	return nil, entity.ErrNotFound
}

func (f *invFake) List(context.Context, string, int, int, string) ([]entity.Invitation, int, error) {
	return nil, 0, nil
}

func (f *invFake) Update(context.Context, *entity.Invitation) error { return nil }

func (f *invFake) Delete(context.Context, string, string) error { return nil }

func (f *invFake) CountByWedding(context.Context, string) (int, error) { return 0, nil }

var _ repository.InvitationRepository = (*invFake)(nil)

type guestFake struct {
	find       *entity.Guest
	findErr    error
	listGuests []entity.Guest
	listErr    error
	updateN    int
	updateErr  error
}

func (g *guestFake) Create(context.Context, *entity.Guest) error { return nil }

func (g *guestFake) FindByID(context.Context, string, string) (*entity.Guest, error) {
	return nil, entity.ErrNotFound
}

func (g *guestFake) FindByNameInInvitation(context.Context, string, string, string) (*entity.Guest, error) {
	if g.findErr != nil {
		return nil, g.findErr
	}
	if g.find != nil {
		return g.find, nil
	}
	return nil, entity.ErrNotFound
}

func (g *guestFake) ListByInvitation(context.Context, string, string) ([]entity.Guest, error) {
	if g.listErr != nil {
		return nil, g.listErr
	}
	return g.listGuests, nil
}

func (g *guestFake) List(context.Context, string, int, int, string, string) ([]entity.Guest, int, error) {
	return nil, 0, nil
}

func (g *guestFake) Update(context.Context, *entity.Guest) error {
	g.updateN++
	return g.updateErr
}

func (g *guestFake) Delete(context.Context, string, string) error { return nil }

func (g *guestFake) CountByWedding(context.Context, string) (int, int, int, int, error) {
	return 0, 0, 0, 0, nil
}

var _ repository.GuestRepository = (*guestFake)(nil)
