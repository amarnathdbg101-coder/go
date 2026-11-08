package deepclone

import (
	"reflect"
)

// Clone creates a deep, independent copy of the input value of type T.
// All nested pointers, maps, slices, and structs are duplicated in new memory.
// It detects and preserves circular pointer references safely without infinite recursion.
func Clone[T any](src T) T {
	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return src
	}

	visited := make(map[uintptr]reflect.Value)
	cloned := cloneRecursive(v, visited)
	return cloned.Interface().(T)
}

// CloneValue creates a deep copy of any interface value.
func CloneValue(src any) any {
	if src == nil {
		return nil
	}
	return Clone(src)
}

func cloneRecursive(v reflect.Value, visited map[uintptr]reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}

		ptr := v.Pointer()
		if existing, found := visited[ptr]; found {
			return existing
		}

		newPtr := reflect.New(v.Elem().Type())
		visited[ptr] = newPtr

		clonedElem := cloneRecursive(v.Elem(), visited)
		newPtr.Elem().Set(clonedElem)
		return newPtr

	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}

		newSlice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		for i := 0; i < v.Len(); i++ {
			newSlice.Index(i).Set(cloneRecursive(v.Index(i), visited))
		}
		return newSlice

	case reflect.Array:
		newArr := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			newArr.Index(i).Set(cloneRecursive(v.Index(i), visited))
		}
		return newArr

	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}

		newMap := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, key := range v.MapKeys() {
			clonedKey := cloneRecursive(key, visited)
			clonedVal := cloneRecursive(v.MapIndex(key), visited)
			newMap.SetMapIndex(clonedKey, clonedVal)
		}
		return newMap

	case reflect.Struct:
		newStruct := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			field := newStruct.Field(i)
			if field.CanSet() {
				field.Set(cloneRecursive(v.Field(i), visited))
			}
		}
		return newStruct

	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		return cloneRecursive(v.Elem(), visited)

	default:
		// Primitives (int, string, bool, float, etc.) are immutable / value types
		return v
	}
}
