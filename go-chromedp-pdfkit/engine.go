package pdfkit

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Engine manages headless browser execution, resource lifecycle, and PDF conversion.
type Engine struct {
	allocatorOpts  []chromedp.ExecAllocatorOption
	defaultTimeout time.Duration
}

// EngineOption allows customization of the Engine.
type EngineOption func(*Engine)

// WithTimeout sets maximum time allowed for rendering and PDF generation.
func WithTimeout(d time.Duration) EngineOption {
	return func(e *Engine) {
		if d > 0 {
			e.defaultTimeout = d
		}
	}
}

// WithExecPath specifies a custom path to Chrome or Edge binary.
func WithExecPath(path string) EngineOption {
	return func(e *Engine) {
		if path != "" {
			e.allocatorOpts = append(e.allocatorOpts, chromedp.ExecPath(path))
		}
	}
}

// WithAdditionalFlags appends custom Chromium command line flags.
func WithAdditionalFlags(flags ...chromedp.ExecAllocatorOption) EngineOption {
	return func(e *Engine) {
		e.allocatorOpts = append(e.allocatorOpts, flags...)
	}
}

// New creates and initializes a PDF generation Engine with production-safe headless Chromium flags.
func New(opts ...EngineOption) *Engine {
	// Production-hardened headless flags: Prevents crashes, GPU deadlocks, and background CPU spikes
	defaultFlags := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", "true"),
		chromedp.Flag("disable-background-networking", "true"),
		chromedp.Flag("disable-background-timer-throttling", "true"),
		chromedp.Flag("disable-breakpad", "true"),
		chromedp.Flag("disable-sync", "true"),
		chromedp.Flag("disable-translate", "true"),
		chromedp.Flag("hide-scrollbars", "true"),
		chromedp.Flag("mute-audio", "true"),
	)

	e := &Engine{
		allocatorOpts:  defaultFlags,
		defaultTimeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// GenerateFromHTML converts raw HTML string into PDF byte buffer.
func (e *Engine) GenerateFromHTML(ctx context.Context, htmlContent string, opts PDFOptions) ([]byte, error) {
	if htmlContent == "" {
		return nil, ErrEmptyHTML
	}

	// Base64 Data URL handles arbitrarily large HTML without path escaping or special character issues
	dataURL := fmt.Sprintf("data:text/html;charset=utf-8;base64,%s", base64.StdEncoding.EncodeToString([]byte(htmlContent)))
	return e.renderAndPrint(ctx, dataURL, opts)
}

// GenerateFromURL navigates to a live web page or local HTTP URL and prints it to PDF.
func (e *Engine) GenerateFromURL(ctx context.Context, targetURL string, opts PDFOptions) ([]byte, error) {
	if targetURL == "" {
		return nil, ErrEmptyURL
	}
	return e.renderAndPrint(ctx, targetURL, opts)
}

// GenerateFromTemplate renders Go html/template with data and converts the result to PDF.
func (e *Engine) GenerateFromTemplate(ctx context.Context, tmplStr string, data any, opts PDFOptions) ([]byte, error) {
	renderedHTML, err := RenderHTML(tmplStr, data)
	if err != nil {
		return nil, err
	}
	return e.GenerateFromHTML(ctx, renderedHTML, opts)
}

// SaveToFile is a convenience helper that writes generated PDF bytes directly to disk.
func SaveToFile(pdfBytes []byte, outputPath string) error {
	return os.WriteFile(outputPath, pdfBytes, 0644)
}

// renderAndPrint executes Chromium actions with guaranteed context cancellation to prevent zombie chrome.exe leaks.
func (e *Engine) renderAndPrint(ctx context.Context, navigateURL string, opts PDFOptions) ([]byte, error) {
	opts.Sanitize()

	// 1. Allocator context (Controls browser process lifecycle)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, e.allocatorOpts...)
	defer cancelAlloc() // Anti-Zombie Pattern: Guarantees chrome.exe is killed on function exit

	// 2. Tab Context
	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	// 3. Execution Timeout Context
	timeoutCtx, cancelTimeout := context.WithTimeout(taskCtx, e.defaultTimeout)
	defer cancelTimeout()

	// Build action chain using chromedp Action[Void] steps
	var steps []chromedp.Action[chromedp.Void]

	steps = append(steps,
		chromedp.Navigate(navigateURL),
		chromedp.WaitReady("body"),
	)

	// Optional wait for specific CSS DOM element (e.g. dynamic charts, barcode images)
	if opts.WaitSelector != "" {
		steps = append(steps, chromedp.WaitVisible(opts.WaitSelector))
	}

	// Optional wait duration for web fonts to load
	if opts.WaitDuration > 0 {
		steps = append(steps, chromedp.Sleep(opts.WaitDuration))
	}

	var pdfBuffer []byte

	// Chrome DevTools Protocol PrintToPDF step
	printStep := func(c context.Context, t *chromedp.Target) (chromedp.Void, error) {
		params := page.PrintToPDFParams{
			Landscape:           ptrBool(opts.Landscape),
			DisplayHeaderFooter: ptrBool(opts.DisplayHeaderFooter),
			PrintBackground:     ptrBool(opts.PrintBackground),
			Scale:               opts.Scale,
			PaperWidth:          opts.PaperWidth,
			PaperHeight:         opts.PaperHeight,
			MarginTop:           ptrFloat64(opts.MarginTop),
			MarginBottom:        ptrFloat64(opts.MarginBottom),
			MarginLeft:          ptrFloat64(opts.MarginLeft),
			MarginRight:         ptrFloat64(opts.MarginRight),
			HeaderTemplate:      opts.HeaderTemplate,
			FooterTemplate:      opts.FooterTemplate,
		}

		res, err := cdp.Call(c, t, page.PrintToPDF, params)
		if err != nil {
			return chromedp.Void{}, err
		}

		pdfBuffer = res.Data
		return chromedp.Void{}, nil
	}

	steps = append(steps, printStep)

	// Execute chromedp steps sequentially
	if err := chromedp.Do(timeoutCtx, steps...); err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, ErrBrowserTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrBrowserFailed, err)
	}

	return pdfBuffer, nil
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrFloat64(f float64) *float64 {
	return &f
}
