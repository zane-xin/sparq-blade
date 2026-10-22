package engine

import "reflect"

// DataFields 表示业务标签中常用的对象字段集合。
type DataFields map[string]any

// Fields 创建一组可继续链式追加的字段集合。
func Fields() DataFields {
	return make(DataFields)
}

// Set 写入一个字段并返回当前字段集合，便于链式构造。
func (f DataFields) Set(key string, value any) DataFields {
	f[key] = value
	return f
}

// Merge 把一组字段合并到当前字段集合。
func (f DataFields) Merge(values map[string]any) DataFields {
	for key, value := range values {
		f[key] = value
	}
	return f
}

// Item 把字段集合转换为单条数据项。
func Item(fields DataFields) any {
	return map[string]any(fields)
}

// Items 用多条数据项快速构造列表结果。
func Items(items ...any) DataTagResult {
	return DataTagResult{Items: items}
}

// Rows 用多条 map、struct 或 slice/array 快速构造列表结果。
func Rows(rows ...any) DataTagResult {
	if len(rows) == 1 {
		if expanded, ok := expandRows(rows[0]); ok {
			return DataTagResult{Items: expanded}
		}
	}

	return DataTagResult{Items: rows}
}

// ValueOf 用单个对象快速构造单值结果。
func ValueOf(value any) DataTagResult {
	return DataTagResult{Value: value}
}

// Object 用单个 map 或 struct 快速构造单对象结果。
func Object(value any) DataTagResult {
	return ValueOf(value)
}

// EmptyResult 返回一个空业务标签结果。
func EmptyResult() DataTagResult {
	return DataTagResult{}
}

func expandRows(value any) ([]any, bool) {
	if value == nil {
		return nil, true
	}

	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return nil, true
		}
		reflected = reflected.Elem()
	}

	switch reflected.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]any, 0, reflected.Len())
		for index := 0; index < reflected.Len(); index++ {
			items = append(items, reflected.Index(index).Interface())
		}
		return items, true
	default:
		return nil, false
	}
}
