package netlib

import (
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
)

type startIoService struct {
	sc *SessionConfig
}

func (sis *startIoService) Done(o *basic.Object) error {

	s := NetModule.newIoService(sis.sc)
	if s != nil {
		NetModule.pool[sis.sc.Id] = s
		s.start()
	}

	return nil
}

func SendStartNetIoService(s *basic.Object, sc *SessionConfig) bool {
	return core.CoreObject().SendCommand(s, &startIoService{sc: sc}, false)
}
