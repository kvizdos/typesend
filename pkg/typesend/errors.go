package typesend

import "errors"

var (
	TypeSendError_INVALID_EMAIL = errors.New("typesend: invalid email format")
	TypeSendError_UTC_MISMATCH  = errors.New("typesend: date must be in UTC")
)
