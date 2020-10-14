package utils

import (
	"testing"

	"code.google.com/p/goprotobuf/proto"
)

type StructD struct {
	Int4 int
}

type StructC struct {
	Int3 int
	DMap map[int]*StructD
	DArr []StructD
	TMap map[int]struct{}
}

type StructB struct {
	StructC
	C        *int32
	IntSlice []*int32
	Map      map[int]string
}
type StructA struct {
	IntValue   int
	StrValue   string
	InnerValue *StructB
}

func TestClone(t *testing.T) {
	a := &StructA{IntValue: 1, StrValue: "test",
		InnerValue: &StructB{
			C:        proto.Int(9),
			IntSlice: []*int32{proto.Int(1), proto.Int(2), proto.Int(3), proto.Int(4), proto.Int(5), proto.Int(6), proto.Int(7), proto.Int(8), proto.Int(9), proto.Int(0)},
			Map:      map[int]string{1: "test", 2: "Hello"},
			StructC:  StructC{Int3: 33, DMap: map[int]*StructD{1: &StructD{Int4: 44}}},
		},
	}
	b := Clone(a).(*StructA)
	//t.Trace(a, b)

	b.InnerValue.IntSlice[0] = proto.Int(99)
	b.InnerValue.IntSlice = b.InnerValue.IntSlice[:7]
	//t.Tracef("%#v %#v %#v\r\n", a, a.InnerValue.IntSlice, *a.InnerValue.C)
	//t.Tracef("%#v %#v %#v\r\n", b, b.InnerValue.IntSlice, *b.InnerValue.C)
	//t.Tracef("%#v\r\n", a.InnerValue.Map)
	//t.Tracef("%#v\r\n", b.InnerValue.Map)
	//t.Tracef("%#v\r\n", a.InnerValue.StructC)
	//t.Tracef("%#v\r\n", b.InnerValue.StructC)
}

func BenchmarkClone(b *testing.B) {
	a := &StructA{IntValue: 1, StrValue: "test",
		InnerValue: &StructB{
			C:        proto.Int(9),
			IntSlice: []*int32{proto.Int(1), proto.Int(2), proto.Int(3), proto.Int(4), proto.Int(5), proto.Int(6), proto.Int(7), proto.Int(8), proto.Int(9), proto.Int(0)},
			Map:      map[int]string{1: "test", 2: "Hello"},
			StructC:  StructC{Int3: 33, DMap: map[int]*StructD{1: &StructD{Int4: 44}}},
		},
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = Clone(a).(*StructA)
	}
	b.StopTimer()
}
