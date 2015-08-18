// Gbp
package netlib

import (
	"errors"

	"code.google.com/p/goprotobuf/proto"
)

var ErrorTypeNotFit = errors.New("packet not proto.Message type")

var Gpb = &GbpEncDecoder{}

type GbpEncDecoder struct {
}

func (this *GbpEncDecoder) Unmarshal(buf []byte, pack interface{}) error {
	if protomsg, ok := pack.(proto.Message); ok {
		err := proto.Unmarshal(buf, protomsg)
		if err != nil {
			return err
		} else {
			return nil
		}
	}

	return ErrorTypeNotFit
}

func (this *GbpEncDecoder) Marshal(pack interface{}) ([]byte, error) {
	if protomsg, ok := pack.(proto.Message); ok {
		return proto.Marshal(protomsg)
	}

	return nil, ErrorTypeNotFit
}

func init() {
	RegisteEncoding(EncodingTypeGPB, Gpb, func(pack interface{}) int {
		if _, ok := pack.(proto.Message); ok {
			return EncodingTypeGPB
		}
		return -1
	})
}
