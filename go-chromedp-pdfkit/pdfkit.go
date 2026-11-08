package pdfkit

import "context"

var defaultEngine = New()

// GenerateFromHTML converts HTML string to PDF using default engine settings.
func GenerateFromHTML(ctx context.Context, htmlContent string, opts PDFOptions) ([]byte, error) {
	return defaultEngine.GenerateFromHTML(ctx, htmlContent, opts)
}

// GenerateFromURL renders a remote or local URL into PDF using default engine settings.
func GenerateFromURL(ctx context.Context, targetURL string, opts PDFOptions) ([]byte, error) {
	return defaultEngine.GenerateFromURL(ctx, targetURL, opts)
}

// GenerateFromTemplate renders Go template with data into PDF using default engine settings.
func GenerateFromTemplate(ctx context.Context, tmplStr string, data any, opts PDFOptions) ([]byte, error) {
	return defaultEngine.GenerateFromTemplate(ctx, tmplStr, data, opts)
}

// GenerateInvoice generates an A4 invoice PDF using the built-in modern invoice template.
func GenerateInvoice(ctx context.Context, data InvoiceData) ([]byte, error) {
	return defaultEngine.GenerateFromTemplate(ctx, DefaultInvoiceTemplate, data, InvoiceOptions())
}
