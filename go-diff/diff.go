package godiff

import (
	"fmt"
	"reflect"
	"strings"
)

// ChangeType represents the nature of state difference.
type ChangeType string

const (
	Added    ChangeType = "ADDED"
	Removed  ChangeType = "REMOVED"
	Modified ChangeType = "MODIFIED"
)

// DiffItem represents an individual field-level state difference.
type DiffItem struct {
	Path     string     `json:"path"`
	Type     ChangeType `json:"type"`
	OldValue any        `json:"old_value,omitempty"`
	NewValue any        `json:"new_value,omitempty"`
}

// DiffResult holds all detected changes between two data structures.
type DiffResult struct {
	Changes []DiffItem `json:"changes"`
}

// HasChanges returns true if any differences were detected.
func (r *DiffResult) HasChanges() bool {
	return len(r.Changes) > 0
}

// Added returns only items that were added in the new state.
func (r *DiffResult) Added() []DiffItem {
	var items []DiffItem
	for _, c := range r.Changes {
		if c.Type == Added {
			items = append(items, c)
		}
	}
	return items
}

// Removed returns only items that were removed from the old state.
func (r *DiffResult) Removed() []DiffItem {
	var items []DiffItem
	for _, c := range r.Changes {
		if c.Type == Removed {
			items = append(items, c)
		}
	}
	return items
}

// Modified returns only items whose values changed between states.
func (r *DiffResult) Modified() []DiffItem {
	var items []DiffItem
	for _, c := range r.Changes {
		if c.Type == Modified {
			items = append(items, c)
		}
	}
	return items
}

// Compare compares two structures, maps, or slices and computes their difference.
func Compare(oldVal, newVal any) (*DiffResult, error) {
	res := &DiffResult{
		Changes: make([]DiffItem, 0),
	}

	vOld := reflect.ValueOf(oldVal)
	vNew := reflect.ValueOf(newVal)

	diffRecursive("", vOld, vNew, res)
	return res, nil
}

func diffRecursive(path string, vOld, vNew reflect.Value, res *DiffResult) {
	// Handle invalid / nil values
	if !vOld.IsValid() && !vNew.IsValid() {
		return
	}

	if !vOld.IsValid() && vNew.IsValid() {
		res.Changes = append(res.Changes, DiffItem{
			Path:     cleanPath(path),
			Type:     Added,
			NewValue: formatValue(vNew),
		})
		return
	}

	if vOld.IsValid() && !vNew.IsValid() {
		res.Changes = append(res.Changes, DiffItem{
			Path:     cleanPath(path),
			Type:     Removed,
			OldValue: formatValue(vOld),
		})
		return
	}

	// Unwrap pointers and interfaces
	vOld = unwrap(vOld)
	vNew = unwrap(vNew)

	if !vOld.IsValid() && !vNew.IsValid() {
		return
	}
	if !vOld.IsValid() && vNew.IsValid() {
		res.Changes = append(res.Changes, DiffItem{
			Path:     cleanPath(path),
			Type:     Added,
			NewValue: formatValue(vNew),
		})
		return
	}
	if vOld.IsValid() && !vNew.IsValid() {
		res.Changes = append(res.Changes, DiffItem{
			Path:     cleanPath(path),
			Type:     Removed,
			OldValue: formatValue(vOld),
		})
		return
	}

	// Check type mismatch
	if vOld.Type() != vNew.Type() {
		res.Changes = append(res.Changes, DiffItem{
			Path:     cleanPath(path),
			Type:     Modified,
			OldValue: formatValue(vOld),
			NewValue: formatValue(vNew),
		})
		return
	}

	switch vOld.Kind() {
	case reflect.Struct:
		diffStruct(path, vOld, vNew, res)
	case reflect.Map:
		diffMap(path, vOld, vNew, res)
	case reflect.Slice, reflect.Array:
		diffSlice(path, vOld, vNew, res)
	default:
		// Primitive comparison
		if !reflect.DeepEqual(vOld.Interface(), vNew.Interface()) {
			res.Changes = append(res.Changes, DiffItem{
				Path:     cleanPath(path),
				Type:     Modified,
				OldValue: formatValue(vOld),
				NewValue: formatValue(vNew),
			})
		}
	}
}

func diffStruct(path string, vOld, vNew reflect.Value, res *DiffResult) {
	t := vOld.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		diffTag := field.Tag.Get("diff")
		if diffTag == "-" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" && diffTag == "" {
			continue
		}

		name := field.Name
		if diffTag != "" {
			name = strings.Split(diffTag, ",")[0]
		} else if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")[0]
			if parts != "" {
				name = parts
			}
		}

		subPath := name
		if path != "" {
			subPath = fmt.Sprintf("%s.%s", path, name)
		}

		diffRecursive(subPath, vOld.Field(i), vNew.Field(i), res)
	}
}

func diffMap(path string, vOld, vNew reflect.Value, res *DiffResult) {
	allKeys := make(map[any]bool)
	for _, k := range vOld.MapKeys() {
		allKeys[k.Interface()] = true
	}
	for _, k := range vNew.MapKeys() {
		allKeys[k.Interface()] = true
	}

	for k := range allKeys {
		kVal := reflect.ValueOf(k)
		oldItem := vOld.MapIndex(kVal)
		newItem := vNew.MapIndex(kVal)

		subPath := fmt.Sprintf("%v", k)
		if path != "" {
			subPath = fmt.Sprintf("%s.%v", path, k)
		}

		diffRecursive(subPath, oldItem, newItem, res)
	}
}

func diffSlice(path string, vOld, vNew reflect.Value, res *DiffResult) {
	lenOld := vOld.Len()
	lenNew := vNew.Len()

	maxLen := lenOld
	if lenNew > maxLen {
		maxLen = lenNew
	}

	for i := 0; i < maxLen; i++ {
		subPath := fmt.Sprintf("%s[%d]", path, i)
		if i >= lenOld {
			res.Changes = append(res.Changes, DiffItem{
				Path:     cleanPath(subPath),
				Type:     Added,
				NewValue: formatValue(vNew.Index(i)),
			})
		} else if i >= lenNew {
			res.Changes = append(res.Changes, DiffItem{
				Path:     cleanPath(subPath),
				Type:     Removed,
				OldValue: formatValue(vOld.Index(i)),
			})
		} else {
			diffRecursive(subPath, vOld.Index(i), vNew.Index(i), res)
		}
	}
}

func unwrap(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func formatValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

func cleanPath(p string) string {
	return strings.TrimPrefix(p, ".")
}
