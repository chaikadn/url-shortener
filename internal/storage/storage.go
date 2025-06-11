package storage

import (
	"errors"

	"github.com/chaikadn/url-shortener/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrURLNotFound       = errors.New("url not found")
	ErrKeyNotFound       = errors.New("key not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrURLAlreadyExists  = errors.New("url already exists")
	ErrKeyAlreadyExists  = errors.New("key already exists")

	ErrLinkAlreadyExists = errors.New("user and url already linked")
)

type UserStorage interface {
	SaveUser(user *model.User) (userID int, err error)
	GetUserByID(userID int) (*model.User, error)
	GetUserByName(username string) (*model.User, error)
}

type URLStorage interface {
	SaveEntry(userID int, entry *model.URLEntry) (entryID int, err error)

	GetEntryByID(entryID int) (*model.URLEntry, error)
	GetEntryByKey(key string) (*model.URLEntry, error)
	GetEntryByURL(originalURL string) (*model.URLEntry, error)

	LinkUserWithEntry(userID int, entryID int) error
	GetUserURLs(userID int) ([]*model.URLEntry, error)

	Ping() error
	Close() error
}
