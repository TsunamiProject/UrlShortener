package db

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrDuplicateURL Error = "URL is already exist"
)
