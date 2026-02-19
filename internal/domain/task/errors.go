package task

import "errors"

var (
	ErrInvalidSort        = errors.New("invalid sort")
	ErrInvalidCursor      = errors.New("invalid cursor")
	ErrInvalidLimit       = errors.New("invalid limit")
	ErrConflict           = errors.New("conflict")
	ErrEmptyTitle         = errors.New("need title")
	ErrTitleTooLong       = errors.New("too long title")
	ErrEmptyDescription   = errors.New("need description")
	ErrDescriptionTooLong = errors.New("too long description")
	ErrInvalidDueOption   = errors.New("invalid due")
	ErrInvalidID          = errors.New("invalid ID")
	ErrNotFound           = errors.New("not found")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrStatusChangeDone   = errors.New("cannot change status of done task")
)
