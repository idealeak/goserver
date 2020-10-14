package utils

import (
	"reflect"
)

//unsupport [Complex64,Complex128,Chan,Func,Interface,UnsafePointer]
func Clone(src interface{}) (dst interface{}) {
	if !isStructPtr(reflect.TypeOf(src)) {
		return nil
	}

	sv := reflect.Indirect(reflect.ValueOf(src))
	if !sv.IsValid() {
		return nil
	}
	st := sv.Type()
	dv := reflect.New(st)
	if !dv.IsValid() {
		return nil
	}
	deepCopy(sv, dv.Elem(), st)
	return dv.Interface()
}

func deepCopy(src, dst reflect.Value, t reflect.Type) {
	switch src.Kind() {
	case reflect.String:
		dst.SetString(src.String())
	case reflect.Bool:
		dst.SetBool(src.Bool())
	case reflect.Float32, reflect.Float64:
		dst.SetFloat(src.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetInt(src.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dst.SetUint(src.Uint())
	case reflect.Map:
		deepCopyMap(src, dst, t)
	case reflect.Array, reflect.Slice:
		deepCopySlice(src, dst, t)
	case reflect.Struct:
		deepCopyStruct(src, dst, t)
	}
}

func deepCopyMap(src, dst reflect.Value, t reflect.Type) {
	for _, key := range src.MapKeys() {
		var nkey, nval reflect.Value
		if key.IsValid() && key.CanSet() {
			if key.Kind() == reflect.Ptr {
				nkey = reflect.New(key.Elem().Type())
			} else {
				nkey = reflect.New(key.Type())
				nkey = reflect.Indirect(nkey)
			}
			s := reflect.Indirect(key)
			d := reflect.Indirect(nkey)
			if s.IsValid() && d.IsValid() {
				tt := s.Type()
				deepCopy(s, d, tt)
			}
		} else {
			nkey = key
		}
		if val := src.MapIndex(key); val.IsValid() && val.CanSet() {
			if val.Kind() == reflect.Ptr {
				nval = reflect.New(val.Elem().Type())
			} else {
				nval = reflect.New(val.Type())
				nval = reflect.Indirect(nval)
			}
			s := reflect.Indirect(val)
			d := reflect.Indirect(nval)
			if s.IsValid() && d.IsValid() {
				tt := s.Type()
				deepCopy(s, d, tt)
			}
		} else {
			nval = val
		}
		dst.SetMapIndex(nkey, nval)
	}
}

func deepCopySlice(src, dst reflect.Value, t reflect.Type) {
	for i := 0; i < src.Len(); i++ {
		sf := src.Index(i)
		df := dst.Index(i)

		if sf.Kind() == reflect.Ptr {
			df = reflect.New(sf.Elem().Type())
			dst.Index(i).Set(df)
		}
		sf = reflect.Indirect(sf)
		df = reflect.Indirect(df)
		if sf.IsValid() && df.IsValid() {
			tt := sf.Type()
			deepCopy(sf, df, tt)
		}
	}
}

func deepCopyStruct(src, dst reflect.Value, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		sv := src.Field(i)
		if sv.CanSet() && sv.IsValid() {
			switch sv.Kind() {
			case reflect.Ptr:
				if !sv.IsNil() {
					dst.Field(i).Set(reflect.New(sv.Elem().Type()))
				}
			case reflect.Array, reflect.Slice:
				if !sv.IsNil() {
					dst.Field(i).Set(reflect.MakeSlice(sv.Type(), sv.Len(), sv.Cap()))
				}
			case reflect.Map:
				if !sv.IsNil() {
					dst.Field(i).Set(reflect.MakeMap(sv.Type()))
				}
			}
			sf := reflect.Indirect(sv)
			df := reflect.Indirect(dst.Field(i))
			if sf.IsValid() && df.IsValid() {
				tt := sf.Type()
				deepCopy(sf, df, tt)
			}
		}
	}
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}
