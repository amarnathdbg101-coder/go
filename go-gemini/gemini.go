// Package gogemini provides a zero-dependency Google Gemini REST SDK featuring
// multi-key round-robin rotation, 429 rate-limit cooldown quarantine, multi-model fallback,
// vision OCR, and structured JSON extraction.
package gogemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Client is the primary interface for interacting with Google Gemini models.
type Client struct {
	cfg        Config
	rotator    *KeyRotator
	httpClient *http.Client
	models     []string
}

// New constructs an initialized Gemini Client.
func New(cfg Config) (*Client, error) {
	rotator, err := NewKeyRotator(cfg.APIKeys, cfg.CooldownDuration)
	if err != nil {
		return nil, err
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	primary := cfg.DefaultModel
	if primary == "" {
		primary = ModelGemini20Flash
	}

	models := []string{primary}
	for _, fb := range cfg.FallbackModels {
		if fb != "" && fb != primary {
			models = append(models, fb)
		}
	}

	return &Client{
		cfg:        cfg,
		rotator:    rotator,
		httpClient: httpClient,
		models:     models,
	}, nil
}

// Part represents a component of a Gemini content turn (text or inline data).
type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

// InlineData carries base64-encoded media for Gemini vision models.
type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// Content defines a single prompt turn.
type Content struct {
	Parts []Part `json:"parts"`
}

// GenerationConfig configures temperature and token limits.
type GenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
}

type geminiPayload struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GenerateText sends a plain text prompt through multi-key rotation and multi-model fallback.
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}

	payload := geminiPayload{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}
	if c.cfg.Temperature != nil || c.cfg.MaxOutputTokens > 0 {
		payload.GenerationConfig = &GenerationConfig{
			Temperature:     c.cfg.Temperature,
			MaxOutputTokens: c.cfg.MaxOutputTokens,
		}
	}

	return c.executeWithFallback(ctx, payload)
}

// GenerateWithVision sends a text prompt along with raw image bytes (JPEG, PNG, WebP) to Gemini Vision.
func (c *Client) GenerateWithVision(ctx context.Context, prompt string, imgBytes []byte, mimeType string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}
	if len(imgBytes) == 0 {
		return "", ErrEmptyPayload
	}
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	b64Image := base64.StdEncoding.EncodeToString(imgBytes)

	payload := geminiPayload{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     b64Image,
						},
					},
				},
			},
		},
	}

	return c.executeWithFallback(ctx, payload)
}

// GenerateJSON executes a prompt and automatically deserializes the output into dest.
// Strips markdown ```json ... ``` code fences automatically before unmarshaling.
func GenerateJSON(ctx context.Context, c *Client, prompt string, dest any) error {
	raw, err := c.GenerateText(ctx, prompt)
	if err != nil {
		return err
	}

	cleanJSON := ExtractJSON(raw)
	if err := json.Unmarshal([]byte(cleanJSON), dest); err != nil {
		return fmt.Errorf("%w: %v (raw text: %q)", ErrJSONExtractionFailed, err, raw)
	}

	return nil
}

var jsonBlockRegex = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")

// ExtractJSON removes markdown formatting code fences from model outputs.
func ExtractJSON(input string) string {
	input = strings.TrimSpace(input)
	matches := jsonBlockRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return input
}

func (c *Client) executeWithFallback(ctx context.Context, payload geminiPayload) (string, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("go-gemini: marshal payload: %w", err)
	}

	var lastErr error

	// Try each model in fallback order
	for _, model := range c.models {
		// Try up to count of API keys
		key, keyErr := c.rotator.GetKey()
		if keyErr != nil {
			return "", keyErr
		}

		result, err := c.doPost(ctx, model, key, bodyBytes)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// If 429 rate limited, mark key in cooldown
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
			c.rotator.MarkRateLimited(key)
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("%w: %v", ErrModelFailed, lastErr)
	}
	return "", ErrModelFailed
}

func (c *Client) doPost(ctx context.Context, model, key string, body []byte) (string, error) {
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("go-gemini: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("go-gemini: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("go-gemini: read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("go-gemini: unmarshal response: %w (status %d)", err, resp.StatusCode)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("gemini api error (status %d %s): %s", geminiResp.Error.Code, geminiResp.Error.Status, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("go-gemini: empty candidate response from model %s", model)
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
