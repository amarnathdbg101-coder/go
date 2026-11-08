package contenttype

import (
	"sort"
	"strconv"
	"strings"
)

// MediaType represents a parsed MIME media type with quality score and parameters.
type MediaType struct {
	Type    string
	Subtype string
	MIME    string // Type/Subtype in lowercase, e.g. "application/json"
	Quality float64
	Params  map[string]string
}

// ParseContentType parses an HTTP Content-Type header string (e.g. "application/json; charset=utf-8").
func ParseContentType(header string) (*MediaType, error) {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return nil, ErrEmptyHeader
	}

	mt, err := parseSingleMediaType(trimmed)
	if err != nil {
		return nil, err
	}
	return mt, nil
}

// ParseAccept parses an HTTP Accept header string and returns media types sorted by quality score descending.
func ParseAccept(header string) []MediaType {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	var result []MediaType

	for _, part := range parts {
		mt, err := parseSingleMediaType(part)
		if err == nil {
			result = append(result, *mt)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Quality != result[j].Quality {
			return result[i].Quality > result[j].Quality
		}
		// Prefer specific over wildcards
		return specificityScore(result[i]) > specificityScore(result[j])
	})

	return result
}

func parseSingleMediaType(raw string) (*MediaType, error) {
	parts := strings.Split(raw, ";")
	mimePart := strings.TrimSpace(parts[0])

	slashIdx := strings.IndexByte(mimePart, '/')
	if slashIdx == -1 {
		return nil, ErrInvalidMediaType
	}

	mainType := strings.ToLower(strings.TrimSpace(mimePart[:slashIdx]))
	subType := strings.ToLower(strings.TrimSpace(mimePart[slashIdx+1:]))

	mt := &MediaType{
		Type:    mainType,
		Subtype: subType,
		MIME:    mainType + "/" + subType,
		Quality: 1.0,
		Params:  make(map[string]string),
	}

	for _, param := range parts[1:] {
		param = strings.TrimSpace(param)
		eqIdx := strings.IndexByte(param, '=')
		if eqIdx == -1 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(param[:eqIdx]))
		val := strings.Trim(strings.TrimSpace(param[eqIdx+1:]), `"`)

		if key == "q" {
			if q, err := strconv.ParseFloat(val, 64); err == nil {
				if q >= 0.0 && q <= 1.0 {
					mt.Quality = q
				}
			}
		} else {
			mt.Params[key] = val
		}
	}

	return mt, nil
}

func specificityScore(mt MediaType) int {
	if mt.Type == "*" {
		return 0
	}
	if mt.Subtype == "*" {
		return 1
	}
	return 2
}

// Negotiate selects the best matching content type from supported options based on the Accept header.
func Negotiate(acceptHeader string, supported []string) (string, bool) {
	if len(supported) == 0 {
		return "", false
	}

	if strings.TrimSpace(acceptHeader) == "" {
		return supported[0], true
	}

	accepted := ParseAccept(acceptHeader)
	if len(accepted) == 0 {
		return supported[0], true
	}

	for _, acc := range accepted {
		if acc.Quality <= 0 {
			continue
		}

		for _, sup := range supported {
			supLower := strings.ToLower(sup)
			if acc.MIME == "*/*" {
				return sup, true
			}
			if acc.Subtype == "*" && strings.HasPrefix(supLower, acc.Type+"/") {
				return sup, true
			}
			if acc.MIME == supLower {
				return sup, true
			}
		}
	}

	return "", false
}

// NegotiateLanguage selects the preferred language tag from supported options based on Accept-Language.
func NegotiateLanguage(acceptLangHeader string, supported []string) (string, bool) {
	if len(supported) == 0 {
		return "", false
	}
	if strings.TrimSpace(acceptLangHeader) == "" {
		return supported[0], true
	}

	parts := strings.Split(acceptLangHeader, ",")
	type langItem struct {
		tag string
		q   float64
	}

	var items []langItem
	for _, p := range parts {
		sub := strings.Split(p, ";")
		tag := strings.ToLower(strings.TrimSpace(sub[0]))
		q := 1.0
		for _, opt := range sub[1:] {
			opt = strings.TrimSpace(opt)
			if strings.HasPrefix(opt, "q=") {
				if val, err := strconv.ParseFloat(strings.TrimPrefix(opt, "q="), 64); err == nil {
					q = val
				}
			}
		}
		if q > 0 {
			items = append(items, langItem{tag: tag, q: q})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].q > items[j].q
	})

	for _, item := range items {
		for _, sup := range supported {
			supLower := strings.ToLower(sup)
			if item.tag == "*" || item.tag == supLower || strings.HasPrefix(supLower, item.tag+"-") {
				return sup, true
			}
		}
	}

	return "", false
}

// NegotiateEncoding selects the preferred content encoding (e.g. "gzip", "br") from supported options.
func NegotiateEncoding(acceptEncHeader string, supported []string) (string, bool) {
	if len(supported) == 0 {
		return "", false
	}
	if strings.TrimSpace(acceptEncHeader) == "" {
		return "", false
	}

	lower := strings.ToLower(acceptEncHeader)
	for _, sup := range supported {
		if strings.Contains(lower, strings.ToLower(sup)) {
			return sup, true
		}
	}

	return "", false
}
