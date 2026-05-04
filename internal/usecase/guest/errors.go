package guest

import "errors"

var (
	ErrGuestStatusTransitionNotAllowed = errors.New("guest: transição de status não permitida")
)
