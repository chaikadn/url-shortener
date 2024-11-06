package util

import "net/url"

func IsValidURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}
