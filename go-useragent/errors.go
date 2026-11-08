package useragent

import "errors"

var (
	// ErrEmptyUserAgent is returned when an empty User-Agent string is supplied.
	ErrEmptyUserAgent = errors.New("useragent: empty user agent string")
)
