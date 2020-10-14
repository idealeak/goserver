package basic

var termAckCmd = &termAckCommand{}

type termAckCommand struct {
}

func (tac *termAckCommand) Done(o *Object) error {
	if o == nil {
		return nil
	}

	if o.termAcks > 0 {
		o.termAcks--

		//  This may be a last ack we are waiting for before termination...
		o.checkTermAcks()
	}

	return nil
}

func SendTermAck(p *Object) bool {
	return p.SendCommand(termAckCmd, false)
}
