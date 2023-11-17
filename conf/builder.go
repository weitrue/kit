package conf

import (
	"fmt"
	"github.com/weitrue/kit/utils/mapping"
	"reflect"
	"strings"
)

func buildFieldsInfo(tp reflect.Type, fullName string) (*fieldInfo, error) {
	tp = mapping.Deref(tp)

	switch tp.Kind() {
	case reflect.Struct:
		return buildStructFieldsInfo(tp, fullName)
	case reflect.Array, reflect.Slice:
		return buildFieldsInfo(mapping.Deref(tp.Elem()), fullName)
	case reflect.Chan, reflect.Func:
		return nil, fmt.Errorf("unsupported type: %s", tp.Kind())
	default:
		return &fieldInfo{
			children: make(map[string]*fieldInfo),
		}, nil
	}
}

func buildStructFieldsInfo(tp reflect.Type, fullName string) (*fieldInfo, error) {
	info := &fieldInfo{
		children: make(map[string]*fieldInfo),
	}

	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		if !field.IsExported() {
			continue
		}

		name := getTagName(field)
		lowerCaseName := toLowerCase(name)
		ft := mapping.Deref(field.Type)
		// flatten anonymous fields
		if field.Anonymous {
			if err := buildAnonymousFieldInfo(info, lowerCaseName, ft,
				getFullName(fullName, lowerCaseName)); err != nil {
				return nil, err
			}
		} else if err := buildNamedFieldInfo(info, lowerCaseName, ft,
			getFullName(fullName, lowerCaseName)); err != nil {
			return nil, err
		}
	}

	return info, nil
}

func buildAnonymousFieldInfo(info *fieldInfo, lowerCaseName string, ft reflect.Type, fullName string) error {
	switch ft.Kind() {
	case reflect.Struct:
		fields, err := buildFieldsInfo(ft, fullName)
		if err != nil {
			return err
		}

		for k, v := range fields.children {
			if err = addOrMergeFields(info, k, v, fullName); err != nil {
				return err
			}
		}
	case reflect.Map:
		elemField, err := buildFieldsInfo(mapping.Deref(ft.Elem()), fullName)
		if err != nil {
			return err
		}

		if _, ok := info.children[lowerCaseName]; ok {
			return newConflictKeyError(fullName)
		}

		info.children[lowerCaseName] = &fieldInfo{
			children: make(map[string]*fieldInfo),
			mapField: elemField,
		}
	default:
		if _, ok := info.children[lowerCaseName]; ok {
			return newConflictKeyError(fullName)
		}

		info.children[lowerCaseName] = &fieldInfo{
			children: make(map[string]*fieldInfo),
		}
	}

	return nil
}

func buildNamedFieldInfo(info *fieldInfo, lowerCaseName string, ft reflect.Type, fullName string) error {
	var finfo *fieldInfo
	var err error

	switch ft.Kind() {
	case reflect.Struct:
		finfo, err = buildFieldsInfo(ft, fullName)
		if err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		finfo, err = buildFieldsInfo(ft.Elem(), fullName)
		if err != nil {
			return err
		}
	case reflect.Map:
		elemInfo, err := buildFieldsInfo(mapping.Deref(ft.Elem()), fullName)
		if err != nil {
			return err
		}

		finfo = &fieldInfo{
			children: make(map[string]*fieldInfo),
			mapField: elemInfo,
		}
	default:
		finfo, err = buildFieldsInfo(ft, fullName)
		if err != nil {
			return err
		}
	}

	return addOrMergeFields(info, lowerCaseName, finfo, fullName)
}

// getTagName get the tag name of the given field, if no tag name, use file.Name.
// field.Name is returned on tags like `json:""` and `json:",optional"`.
func getTagName(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup(jsonTagKey); ok {
		if pos := strings.IndexByte(tag, jsonTagSep); pos >= 0 {
			tag = tag[:pos]
		}

		tag = strings.TrimSpace(tag)
		if len(tag) > 0 {
			return tag
		}
	}

	return field.Name
}

func toLowerCase(s string) string {
	return strings.ToLower(s)
}

func addOrMergeFields(info *fieldInfo, key string, child *fieldInfo, fullName string) error {
	if prev, ok := info.children[key]; ok {
		if child.mapField != nil {
			return newConflictKeyError(fullName)
		}

		if err := mergeFields(prev, key, child.children, fullName); err != nil {
			return err
		}
	} else {
		info.children[key] = child
	}

	return nil
}

func mergeFields(prev *fieldInfo, key string, children map[string]*fieldInfo, fullName string) error {
	if len(prev.children) == 0 || len(children) == 0 {
		return newConflictKeyError(fullName)
	}

	// merge fields
	for k, v := range children {
		if _, ok := prev.children[k]; ok {
			return newConflictKeyError(fullName)
		}

		prev.children[k] = v
	}

	return nil
}

func getFullName(parent, child string) string {
	if len(parent) == 0 {
		return child
	}

	return strings.Join([]string{parent, child}, ".")
}
