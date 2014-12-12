// binary
package netlib

import (
	"bytes"
	"encoding/binary"
)

var Bcd = &BinaryEncDecoder{}

type BinaryEncDecoder struct {
}

func (this *BinaryEncDecoder) Unmarshal(buf []byte, pack interface{}) error {
	return binary.Read(bytes.NewReader(buf), binary.LittleEndian, pack)
}

func (this *BinaryEncDecoder) Marshal(pack interface{}) ([]byte, error) {
	writer := bytes.NewBuffer(nil)
	err := binary.Write(writer, binary.LittleEndian, pack)
	return writer.Bytes(), err
}

func init() {
	RegisteEncoding(EncodingTypeBinary, Bcd, func(pack interface{}) int {
		if _, ok := pack.([]byte); ok {
			return EncodingTypeBinary
		}
		return -1
	})
}
