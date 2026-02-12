package phone

import "errors"

// ErrInvalidPhone is returned when a phone number cannot be parsed or is invalid.
var ErrInvalidPhone = errors.New("invalid phone number")
