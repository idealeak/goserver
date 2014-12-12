// Gob
package netlib

import (
	"bytes"
	"encoding/gob"
)

var Gob = &GobEncDecoder{}

type GobEncDecoder struct {
}

func (this *GobEncDecoder) Unmarshal(buf []byte, pack interface{}) error {
	network := bytes.NewBuffer(buf)
	// Create a decoder and receive a value.
	dec := gob.NewDecoder(network)
	err := dec.Decode(pack)
	if err != nil {
		return err
	}

	return nil
}

func (this *GobEncDecoder) Marshal(pack interface{}) ([]byte, error) {
	var network bytes.Buffer // Stand-in for the network.

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&network)
	err := enc.Encode(pack)
	if err != nil {
		return nil, err
	}

	return network.Bytes(), nil
}

func init() {
	RegisteEncoding(EncodingTypeGob, Gob, func(pack interface{}) int {
		return EncodingTypeGob
	})
}
