// nil
package netlib

var Nil = &NilEncDecoder{}

type NilEncDecoder struct {
}

func (this *NilEncDecoder) Unmarshal(buf []byte, pack interface{}) error {
	return nil
}

func (this *NilEncDecoder) Marshal(pack interface{}) ([]byte, error) {
	if binarymsg, ok := pack.([]byte); ok {
		return binarymsg, nil
	}

	return nil, ErrorTypeNotFit
}

func init() {
	RegisteEncoding(EncodingTypeNil, Nil, nil)
}
