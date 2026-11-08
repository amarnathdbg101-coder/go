package queryfilter

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Operator represents SQL comparison operators.
type Operator string

const (
	OpEq    Operator = "="
	OpNeq   Operator = "!="
	OpGt    Operator = ">"
	OpGte   Operator = ">="
	OpLt    Operator = "<"
	OpLte   Operator = "<="
	OpLike  Operator = "LIKE"
	OpILike Operator = "ILIKE"
	OpIn    Operator = "IN"
)

var allowedOps = map[string]Operator{
	"eq":    OpEq,
	"ne":    OpNeq,
	"gt":    OpGt,
	"gte":   OpGte,
	"lt":    OpLt,
	"lte":   OpLte,
	"like":  OpLike,
	"ilike": OpILike,
	"in":    OpIn,
}

// Config configures whitelist boundaries for URL query filtering and sorting.
type Config struct {
	AllowedFilters map[string]bool
	AllowedSorts   map[string]bool
	DefaultSort    string
	DefaultLimit   int
	MaxLimit       int
}

// SQLResult holds the generated parameterized clauses and arguments.
type SQLResult struct {
	WhereClause   string
	OrderByClause string
	Limit         int
	Offset        int
	Args          []any
}

// Parser parses URL query params into SQL clauses.
type Parser struct {
	cfg Config
}

// New creates a new Parser with the configured whitelist.
func New(cfg Config) *Parser {
	if cfg.AllowedFilters == nil {
		cfg.AllowedFilters = make(map[string]bool)
	}
	if cfg.AllowedSorts == nil {
		cfg.AllowedSorts = make(map[string]bool)
	}
	if cfg.DefaultLimit <= 0 {
		cfg.DefaultLimit = 20
	}
	if cfg.MaxLimit <= 0 {
		cfg.MaxLimit = 100
	}

	return &Parser{cfg: cfg}
}

// Parse extracts filters, sorts, and pagination from url.Values into SQL parameterized clauses.
func (p *Parser) Parse(values url.Values, startParamIdx int) (*SQLResult, error) {
	if values == nil {
		return nil, ErrNilURLValues
	}
	if startParamIdx <= 0 {
		startParamIdx = 1
	}

	res := &SQLResult{
		Limit:  p.cfg.DefaultLimit,
		Offset: 0,
	}

	var whereClauses []string
	paramCounter := startParamIdx

	// 1. Process Filters
	for rawKey, vals := range values {
		if len(vals) == 0 {
			continue
		}
		val := vals[0]

		// Check pagination keywords
		if rawKey == "limit" || rawKey == "page" || rawKey == "offset" || rawKey == "sort" {
			continue
		}

		// Parse field and operator (e.g. "age[gte]" -> field="age", op="gte")
		field, op := parseKeyAndOp(rawKey)

		// Whitelist check
		if !p.cfg.AllowedFilters[field] {
			return nil, fmt.Errorf("%w: '%s'", ErrDisallowedField, field)
		}

		sqlOp, ok := allowedOps[op]
		if !ok {
			return nil, fmt.Errorf("%w: '%s'", ErrInvalidOperator, op)
		}

		if sqlOp == OpIn {
			// e.g. "active,pending,suspended"
			inVals := strings.Split(val, ",")
			var inPlaceholders []string
			for _, iv := range inVals {
				inPlaceholders = append(inPlaceholders, fmt.Sprintf("$%d", paramCounter))
				res.Args = append(res.Args, strings.TrimSpace(iv))
				paramCounter++
			}
			whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", field, strings.Join(inPlaceholders, ", ")))
		} else {
			placeholder := fmt.Sprintf("$%d", paramCounter)
			whereClauses = append(whereClauses, fmt.Sprintf("%s %s %s", field, sqlOp, placeholder))
			res.Args = append(res.Args, val)
			paramCounter++
		}
	}

	if len(whereClauses) > 0 {
		res.WhereClause = strings.Join(whereClauses, " AND ")
	}

	// 2. Process Sort
	sortParam := values.Get("sort")
	if sortParam == "" {
		sortParam = p.cfg.DefaultSort
	}

	if sortParam != "" {
		sortFields := strings.Split(sortParam, ",")
		var orderBys []string

		for _, sf := range sortFields {
			sf = strings.TrimSpace(sf)
			if sf == "" {
				continue
			}

			direction := "ASC"
			field := sf
			if strings.HasPrefix(sf, "-") {
				direction = "DESC"
				field = strings.TrimPrefix(sf, "-")
			}

			if !p.cfg.AllowedSorts[field] {
				return nil, fmt.Errorf("%w: sort field '%s'", ErrDisallowedField, field)
			}

			orderBys = append(orderBys, fmt.Sprintf("%s %s", field, direction))
		}

		if len(orderBys) > 0 {
			res.OrderByClause = strings.Join(orderBys, ", ")
		}
	}

	// 3. Process Pagination (limit, page/offset)
	if limitStr := values.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > p.cfg.MaxLimit {
				l = p.cfg.MaxLimit
			}
			res.Limit = l
		}
	}

	if offsetStr := values.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			res.Offset = o
		}
	} else if pageStr := values.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 1 {
			res.Offset = (page - 1) * res.Limit
		}
	}

	return res, nil
}

func parseKeyAndOp(key string) (string, string) {
	if strings.Contains(key, "[") && strings.HasSuffix(key, "]") {
		idx := strings.Index(key, "[")
		field := key[:idx]
		op := key[idx+1 : len(key)-1]
		return field, op
	}
	return key, "eq"
}
