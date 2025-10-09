package storage

type NotFoundError struct {
	Msg string
}

func NewNotFoundError(Msg string) *NotFoundError {
	return &NotFoundError{Msg: Msg}
}

func (e *NotFoundError) Error() string {
	return "По запросу не найден ни один обьект: " + e.Msg
}
