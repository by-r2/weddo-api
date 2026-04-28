package rsvp

import "errors"

var (
	// ErrInvitationNotFound indica que code não corresponde a um convite do tenant.
	ErrInvitationNotFound = errors.New("rsvp: invitation not found")
	// ErrGuestNotFoundOnInvitation indica que o nome não existe entre os convidados daquele convite.
	ErrGuestNotFoundOnInvitation = errors.New("rsvp: guest not found on invitation")
	// ErrGuestStatusTransitionNotAllowed indica que o convidado está em estado terminal e não pode confirmar.
	ErrGuestStatusTransitionNotAllowed = errors.New("rsvp: guest status transition not allowed")
)
