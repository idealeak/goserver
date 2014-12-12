package utils

import (
	"reflect"
	"sort"
)

func AppendEmptySliceField(slice reflect.Value) reflect.Value {
	newField := reflect.Zero(slice.Type().Elem())
	return reflect.Append(slice, newField)
}

func SetSliceLengh(slice reflect.Value, length int) reflect.Value {
	if length > slice.Len() {
		for i := slice.Len(); i < length; i++ {
			slice = AppendEmptySliceField(slice)
		}
	} else if length < slice.Len() {
		slice = slice.Slice(0, length)
	}

	return slice
}

func DeleteEmptySliceElementsVal(sliceVal reflect.Value) reflect.Value {
	if sliceVal.Kind() != reflect.Slice {
		panic("Argument is not a slice: " + sliceVal.String())
	}
	zeroVal := reflect.Zero(sliceVal.Type().Elem())
	for i := 0; i < sliceVal.Len(); i++ {
		elemVal := sliceVal.Index(i)
		if reflect.DeepEqual(elemVal.Interface(), zeroVal.Interface()) {
			before := sliceVal.Slice(0, i)
			after := sliceVal.Slice(i+1, sliceVal.Len())
			sliceVal = reflect.AppendSlice(before, after)
			i--
		}
	}
	return sliceVal
}

func DeleteEmptySliceElements(slice interface{}) interface{} {
	return DeleteEmptySliceElementsVal(reflect.ValueOf(slice)).Interface()
}

func DeleteSliceElementVal(sliceVal reflect.Value, idx int) reflect.Value {
	if idx < 0 || idx >= sliceVal.Len() {
		return sliceVal
	}
	before := sliceVal.Slice(0, idx)
	after := sliceVal.Slice(idx+1, sliceVal.Len())
	sliceVal = reflect.AppendSlice(before, after)
	return sliceVal
}

func DeleteSliceElement(slice interface{}, idx int) interface{} {
	return DeleteSliceElementVal(reflect.ValueOf(slice), idx).Interface()
}

// Implements sort.Interface
type SortableInterfaceSlice struct {
	Slice    []interface{}
	LessFunc func(a, b interface{}) bool
}

func (self *SortableInterfaceSlice) Len() int {
	return len(self.Slice)
}

func (self *SortableInterfaceSlice) Less(i, j int) bool {
	return self.LessFunc(self.Slice[i], self.Slice[j])
}

func (self *SortableInterfaceSlice) Swap(i, j int) {
	self.Slice[i], self.Slice[j] = self.Slice[j], self.Slice[i]
}

func (self *SortableInterfaceSlice) Sort() {
	sort.Sort(self)
}

func SortInterfaceSlice(slice []interface{}, lessFunc func(a, b interface{}) bool) {
	sortable := SortableInterfaceSlice{slice, lessFunc}
	sortable.Sort()
}
