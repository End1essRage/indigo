package storage

type NotFoundError struct {
	Msg string
}

func NewNotFoundError(Msg string) error {
	return &NotFoundError{Msg: Msg}
}

func (e *NotFoundError) Error() string {
	return e.Msg
}
