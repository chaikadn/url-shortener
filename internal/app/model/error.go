package model

import "errors"

var (
	ErrLongURLConflict  = errors.New("long url conflict")
	ErrShortURLConflict = errors.New("short url conflict")
	ErrNotFound         = errors.New("url not found")
	ErrEmptyBatch       = errors.New("empty batch")
)
