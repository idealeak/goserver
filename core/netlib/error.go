// error
package netlib

type NetLibParamError struct {
	Src   string
	Param string
}

func (self *NetLibParamError) Error() string {
	return "Invalid Parameter: " + self.Src + self.Param
}

func newNetLibParamError(src, param string) *NetLibParamError {
	return &NetLibParamError{Src: src, Param: param}
}
