package entity

type Status int

const (
	StatusUnknown Status = iota
	StatusNew
	StatusSuccess
	StatusError
)

const (
	statusUnknown = "unknown"
	statusNew     = "new"
	statusSuccess = "success"
	statusError   = "error"
)

// NewStatus converts a string to Status. Unrecognized values become StatusUnknown.
func NewStatus(s string) Status {
	switch s {
	case statusNew:
		return StatusNew
	case statusSuccess:
		return StatusSuccess
	case statusError:
		return StatusError
	default:
		return StatusUnknown
	}
}

// ParseStatus is a strict version of NewStatus: unrecognized values return ErrStatusInvalid.
func ParseStatus(s string) (Status, error) {
	st := NewStatus(s)
	if st == StatusUnknown && s != statusUnknown {
		return StatusUnknown, ErrStatusInvalid
	}

	return st, nil
}

func (s Status) String() string {
	switch s {
	case StatusNew:
		return statusNew
	case StatusSuccess:
		return statusSuccess
	case StatusError:
		return statusError
	case StatusUnknown:
		return statusUnknown
	default:
		return statusUnknown
	}
}
