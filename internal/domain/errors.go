package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrConflict          = errors.New("conflict")
	ErrAlreadyReserved   = errors.New("already reserved")
	ErrNoAssignment      = errors.New("drone has no assigned order")
	ErrDroneBusy         = errors.New("drone is busy")
	ErrDroneBroken       = errors.New("drone is broken")
	ErrUnauthenticated   = errors.New("unauthenticated")
)
