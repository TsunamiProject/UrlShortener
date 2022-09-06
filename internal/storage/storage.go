package storage

type Storage interface {
	Read(url string) (string, error)
	Write(b []byte, cookieValue string) (string, error)
	ReadAll(authCookie string) (string, error)
	Delete(authCookie string, deleteList []string) error
	Batch(b []byte, cookieValue string) (string, error)
	IsOk() error
}

//int - смешение уровней абстракций - storage должен быть изолирован
