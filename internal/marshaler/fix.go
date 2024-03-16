package marshaler

import "reflect"

func fixProto(v interface{}) interface{} {
	t := reflect.ValueOf(v)
	if t.Elem().Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.IsNil() {
		t.Set(reflect.New(t.Type().Elem()))
	}
	v = t.Interface()
	return v
}
