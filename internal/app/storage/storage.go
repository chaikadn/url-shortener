package storage

type Storage interface {
	Add(longURL string) (shortURL string, err error)
	Get(shortURL string) (longURL string, err error)
}
