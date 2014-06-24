package transact

import (
	"github.com/idealeak/goserver/core/basic"
)

type transactResumeCommand struct {
	tnode *TransNode
}

func (trc *transactResumeCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	trc.tnode.checkExeOver()
	return nil
}

func SendTranscatResume(s *basic.Object, tnode *TransNode) bool {
	return tnode.ownerObj.SendCommand(s, &transactResumeCommand{tnode: tnode}, true)
}
