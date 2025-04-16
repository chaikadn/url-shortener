package validation

import (
	"errors"
	"fmt"
	"net/url"
)

var ErrInvalidURL = errors.New("invalid url")

func ValidateURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("%w: %s", ErrInvalidURL, rawURL)
	}
	return nil
}
