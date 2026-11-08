package pdfkit

import "errors"

// Common domain errors for go-chromedp-pdfkit.
var (
	ErrEmptyHTML       = errors.New("pdfkit: HTML content must not be empty")
	ErrEmptyURL        = errors.New("pdfkit: target URL must not be empty")
	ErrBrowserTimeout  = errors.New("pdfkit: headless browser execution timed out")
	ErrBrowserFailed   = errors.New("pdfkit: chromedp execution failed")
	ErrInvalidOptions  = errors.New("pdfkit: invalid PDF layout options")
	ErrTemplateRender  = errors.New("pdfkit: failed to render HTML template")
)
